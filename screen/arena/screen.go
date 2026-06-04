package arena

import (
	"fmt"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/ability"
	"github.com/ognev-dev/goplease-ebitengine-client/config"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
	"golang.org/x/image/colornames"
)

const (
	abilityCardSize = 64 // ability card size in the footer panel
	unitCardSize    = 64 // unit card size in the hand and queue panel
	unitIconSize    = 54 // unit portrait size rendered on the board hex

	unitStunnedPic = "knockout.png"
	unitKilledPic  = "dead-head.png"

	headerH = 80
	statusH = 32
	footerH = 90
)

// Screen is the main arena game screen.
// It owns all game state visible to the local player and orchestrates
// server communication, UI layout, board rendering, and animations.
type Screen struct {
	server ws.Client
	ui     *ebitenui.UI

	// Game state received from the server.
	board              ds.Board
	roomID             string
	player             ds.Player
	opponentName       string
	isMyTurn           bool
	unitsQueue         []*ds.Unit
	activeUnitID       string
	activeUnitIndex    int
	prevActiveUnitID   string
	roundNumber        int
	unitPlacedThisTurn bool
	queueIn            bool
	unitPanelIn        bool

	// Board rendering.
	safeZoneCells    []*DropZoneCell     // safe-zone cells that accept unit drops
	sortedCells      []*ui.HexCellWidget // board cells sorted by (R, Q) for deterministic overlay render order
	boardCellWidgets map[ds.HexCoord]*ui.HexCellWidget

	// UI widget references used for dynamic updates.
	unitCards     map[string]*widget.Container
	headerRef     *widget.Container
	footerRef     *widget.Container
	queuePanelRef *widget.Container
	unitPanelRef  *widget.Container
	statusBarRef  *widget.Container
	nextActionBtn *widget.Button
	statusLabel   *widget.Text

	// Ability targeting state.
	abilityPanelRef            *widget.Container
	abilityPanelIn             bool
	abilityHighlightCells      []ds.HexCoord // hex coords currently highlighted for ability range
	selectedAbility            *ability.Ability
	selectedAbilityCard        *widget.Container
	selectedAbilityCardColor   color.Color
	selectedAbilityIcon        *widget.Graphic
	selectedAbilityActiveColor color.Color

	// Pulse animation state.
	pulseHexWidgets       []*ui.HexCellWidget // board hex cells that pulse (active unit)
	pulseWidgets          []*widget.Container // other UI widgets that pulse (queue cards, ability card, end turn button)
	pulseTick             float64
	endTurnBtnPulseActive bool

	// Dev panel — only rendered when DevMode.Enabled.
	devPanelRef       *widget.Container
	devPanelBody      *widget.Container
	devLoadList       *widget.Container
	devScenarioList   *widget.Container
	devPanelMinimized bool

	// if dev mode enabled you can see hex coords by holding Alt
	showDevCoordinates bool

	// Movement and selection state.
	selectedUnitID    string        // unit currently selected for movement; empty means none
	reachableCells    []ds.HexCoord // precomputed reachable positions for selectedUnitID
	activeUnitMoved   bool          // true once the active unit has moved this turn
	unitMoveAnimQueue [][]unitMoveAnimAction

	activeFxAnims []*ActiveFxAnim
	// delayedActions holds pending actions scheduled to run after a fixed number of frames.
	delayedActions []delayedAction

	// ready is set to true when the server responds with phase unit_placement,
	// meaning the match has started and the local player may interact.
	ready bool

	// firstDrawn is set after the first Draw call so that ready_to_play
	// is sent to the server exactly once, after the UI is fully rendered.
	firstDrawn bool

	// pendingScreen is set when a screen transition should occur on the next Update.
	pendingScreen game.Screen

	// roundBanner holds the state of the round announcement overlay.
	// nil when no banner is active.
	roundBanner *newRoundBannerState

	// timerBar holds the state of the turn timer progress bar.
	// nil when no timer is active.
	timerBar        *timerBarState
	turnTimeSeconds int // 0 = timer disabled

	floatingTexts  []*floatingText
	pendingVisuals *pendingVisuals
	// pendingDrawOps holds draw operations queued by ProgramFx to be executed in Draw.
	pendingDrawOps []pendingDrawOp
}

// NewScreen constructs a fully initialised arena Screen from a server snapshot.
func NewScreen(snap ds.GameSnapshot, server ws.Client) game.Screen {
	s := &Screen{
		server:          server,
		board:           snap.Board,
		roomID:          snap.RoomID,
		player:          snap.Player,
		opponentName:    snap.OpponentName,
		unitsQueue:      snap.UnitsQueue,
		activeUnitID:    snap.ActiveUnitID,
		roundNumber:     snap.Round,
		unitCards:       make(map[string]*widget.Container),
		turnTimeSeconds: snap.TurnTimeSeconds,
	}

	initDropPointAnim()
	s.setupUI()
	s.restoreBoardVisuals()
	s.rebuildQueuePanel()

	return s
}

// Update processes server messages, handles input, and advances all animations.
// Implements game.Screen.
func (s *Screen) Update(g *game.Game) (game.Screen, error) {
	// Drain all pending server messages before updating game logic.
	for {
		select {
		case msg := <-g.Server.Inbox():
			s.handleServerMessage(msg)
		default:
			goto done
		}
	}
done:

	s.updateDelayedActions()
	s.updateNewRoundBanner()
	s.updateFloatingTexts()

	// Handle hex cell clicks manually since HexCellWidget uses custom hit testing.
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		for coord, cell := range s.boardCellWidgets {
			if cell.HitTest(mx, my) {
				s.onCellClicked(coord)
				break
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.selectedAbility != nil {
			s.cancelAbilitySelection()
		}
	}

	if config.Get().DevMode.Enabled {
		s.showDevCoordinates = ebiten.IsKeyPressed(ebiten.KeyAlt)
	} else {
		s.showDevCoordinates = false
	}

	s.updatePulse()
	s.updateDropZoneAnim()
	s.updateMoveAnimations()
	s.updateFxAnims()
	s.ui.Update()

	if s.pendingScreen != nil {
		return s.pendingScreen, nil
	}

	s.updateTurnTimer()
	return s, nil
}

// Draw renders the screen. Hex cells are drawn via PostRenderHook inside
// s.ui.Draw so they appear above the UI background but below EbitenUI windows
// (drag cards, tooltips). The movement animation is drawn as the topmost layer.
// Implements game.Screen.
func (s *Screen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)

	s.drawActiveFxAnims(screen)

	// Movement animation is rendered above everything including EbitenUI windows.
	if len(s.unitMoveAnimQueue) > 0 {
		for _, action := range s.unitMoveAnimQueue[0] {
			x, y := action.anim.currentPos()

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, y)

			screen.DrawImage(action.anim.img, op)
		}
	}

	// Send ready_to_play once after the first complete draw so the server
	// knows the client is ready to receive game events.
	if !s.firstDrawn {
		s.firstDrawn = true
		s.server.Send(ws.OutMessage{Action: ws.ReadyToPlay})
	}

	s.drawRoundBanner(screen)
	s.drawFloatingTexts(screen)

	for _, d := range s.pendingDrawOps {
		screen.DrawImage(d.img, d.op)
	}
	s.pendingDrawOps = nil
}

// updatePulse advances the sinusoidal pulse animation for highlighted hex cells
// and queue cards. Early-returns if nothing is currently pulsing.
func (s *Screen) updatePulse() {
	if len(s.pulseHexWidgets) == 0 && len(s.pulseWidgets) == 0 &&
		!s.endTurnBtnPulseActive && s.selectedAbilityCard == nil {
		return
	}

	s.pulseTick += 0.05
	t := (math.Sin(s.pulseTick) + 1) / 2
	c := ui.LerpColor(unitPulseColor1, unitPulseColor2, t)

	for _, w := range s.pulseHexWidgets {
		w.SetColor(c)
	}
	for _, w := range s.pulseWidgets {
		w.SetBackgroundImage(image.NewNineSliceColor(c))
	}
	if s.endTurnBtnPulseActive && s.nextActionBtn != nil {
		s.pulseEndTurnBtn(t)
	}

	// Pulse border on the selected ability card.
	if s.selectedAbilityCard != nil {
		borderColor := ui.LerpColor(abilitySelectedPulseColor1, abilitySelectedPulseColor2, t)
		s.selectedAbilityCard.SetBackgroundImage(image.NewBorderedNineSliceColor(
			s.selectedAbilityActiveColor,
			borderColor,
			3,
		))
	}
}

// updateDropZoneAnim advances the drop-arrow animation and syncs the current
// frame to all active drop zone cells.
func (s *Screen) updateDropZoneAnim() {
	animDropArrow.Update()
	for _, sc := range s.safeZoneCells {
		if sc.activeGraphic != nil {
			sc.activeGraphic = animDropArrow.CurrentFrame
		}
	}
}

// setupUI builds the full EbitenUI widget tree and registers the PostRenderHook
// that draws hex cells in the correct layer order.
func (s *Screen) setupUI() {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bodyBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	s.headerRef = s.createHeader()
	s.footerRef = s.createFooter()
	board := s.createBoardContainer()
	statusBar := s.createStatusBar()

	root.AddChild(board)
	root.AddChild(s.headerRef)
	root.AddChild(statusBar)
	root.AddChild(s.footerRef)
	s.setupDevPanel(root)

	s.ui = &ebitenui.UI{
		Container: root,
		// PostRenderHook is called after the main container renders but before
		// EbitenUI windows (drag card, tooltips), giving us the correct layer order:
		// hex fills → grid → unit portraits → drop zone FX → HUD badges → FX effects.
		PostRenderHook: func(screen *ebiten.Image) {
			// Layer 1: hex polygon fills (background color).
			for _, cell := range s.boardCellWidgets {
				cell.RenderFill(screen)
			}
			// Layer 2: grid stroke drawn as a single path to avoid double-width edges.
			s.renderGrid(screen)
			// Layer 3: unit portraits — sorted for deterministic overlap at hex borders.
			for _, cell := range s.sortedCells {
				cell.RenderUnitLayer(screen)
			}
			// Layer 4: drop zone arrow animations.
			for _, sc := range s.safeZoneCells {
				sc.RenderAnim(screen)
			}
			// Layer 5: HUD badges (hp, shield, move indicator).
			for _, cell := range s.sortedCells {
				cell.RenderHUDLayer(screen)
			}
			// Layer 6: FX (damage numbers, attack effects).
			for _, cell := range s.sortedCells {
				cell.RenderFXLayer(screen)
			}

			if s.showDevCoordinates {
				s.drawCellCoordinates(screen)
			}

			// ...
			// Dev panel rendered on top of hex layer.
			if s.devPanelRef != nil {
				s.devPanelRef.Render(screen)
			}
			// Timer bar — rendered above hex layers but below EbitenUI tooltips.
			s.drawTurnTimer(screen)
		},
	}
}

// renderGrid draws the hex grid as a single combined stroke path.
// Using one path avoids double-width edges where adjacent hex strokes overlap.
func (s *Screen) renderGrid(screen *ebiten.Image) {
	var path vector.Path
	for _, cell := range s.boardCellWidgets {
		cell.AppendHexPath(&path)
	}

	var opts vector.DrawPathOptions
	opts.AntiAlias = true
	opts.ColorScale.ScaleWithColor(boardGridColor)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: 1}, &opts)
}

// setStatus updates the status bar label text by recreating the widget to force layout centering.
func (s *Screen) setStatus(text string) {
	if s.statusBarRef == nil {
		return
	}

	// Clear previous text widget to reset cached MinSize dimensions.
	s.statusBarRef.RemoveChildren()

	tf := ui.TextFace(18)
	s.statusLabel = widget.NewText(
		widget.TextOpts.Text(text, &tf, statusBarTextColor),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	s.statusBarRef.AddChild(s.statusLabel)
}

// updateActiveUnitStatusLabel sets the status bar text based on what the active unit can still do.
func (s *Screen) updateActiveUnitStatusLabel() {
	u := s.unitByID(s.activeUnitID)
	if u == nil {
		return
	}

	canMove := !s.activeUnitMoved
	canAct := u.CurrentAP > 0

	var status string
	switch {
	case canMove && canAct:
		status = fmt.Sprintf("%s can move and use an ability", u.Name)
	case canMove:
		status = fmt.Sprintf("%s can move", u.Name)
	case canAct:
		status = fmt.Sprintf("%s can use an ability", u.Name)
	default:
		status = fmt.Sprintf("%s may end turn", u.Name)
	}

	s.setStatus(status)
}

// takeSnapshot captures the current game state into a GameSnapshot.
// Used before navigating away from the screen so state can be restored.
func (s *Screen) takeSnapshot() ds.GameSnapshot {
	return ds.GameSnapshot{
		RoomID:       s.roomID,
		Board:        s.board,
		Player:       s.player,
		OpponentName: s.opponentName,
		UnitsQueue:   s.unitsQueue,
		ActiveUnitID: s.activeUnitID,
		Round:        s.roundNumber,
	}
}

// restoreSnapshot creates a new Screen from a previously captured snapshot.
func (s *Screen) restoreSnapshot(snap ds.GameSnapshot) game.Screen {
	return NewScreen(snap, s.server)
}

// restoreBoardVisuals re-renders unit cards on the board after a snapshot restore.
// Called once during NewScreen before the first Draw.
func (s *Screen) restoreBoardVisuals() {
	for pos, cell := range s.board.Cells {
		if cell == nil || cell.Unit == nil {
			continue
		}

		u := cell.Unit

		w := s.boardCellWidgets[pos]
		if w == nil {
			continue
		}

		bg := unitFriendlyBgColor
		if u.IsOpponent {
			bg = unitEnemyBgColor
		}

		w.SetColor(bg)
		buildBoardCard(w, u, false)
	}
}

// updateTurnControls updates the Next button label and status bar
// based on what the active unit can still do.
// If the unit has exhausted all actions, the turn ends automatically.
func (s *Screen) updateTurnControls() {
	u := s.unitByID(s.activeUnitID)
	if u == nil {
		return
	}

	canMove := !s.activeUnitMoved
	canAct := u.CurrentAP > 0

	if !canMove && !canAct {
		s.server.Send(ws.OutMessage{Action: ws.EndTurnAction})
		return
	}

	s.updateNextActionLabel()
	s.updateActiveUnitStatusLabel()
}

type pendingDrawOp struct {
	img *ebiten.Image
	op  *ebiten.DrawImageOptions
}

// drawOnTop queues an image to be drawn on top of everything in the next Draw call.
func (s *Screen) drawOnTop(img *ebiten.Image, op *ebiten.DrawImageOptions) {
	s.pendingDrawOps = append(s.pendingDrawOps, pendingDrawOp{img: img, op: op})
}

func (s *Screen) drawCellCoordinates(screen *ebiten.Image) {
	for coord, w := range s.boardCellWidgets {
		rect := w.GetWidget().Rect

		centerX := float64(rect.Min.X + rect.Dx()/2)
		centerY := float64(rect.Min.Y + rect.Dy()/2)

		face := ui.TextFace(16)

		op := &text.DrawOptions{}
		op.ColorScale.ScaleWithColor(colornames.Black)
		op.GeoM.Translate(centerX-16, centerY-6)

		text.Draw(screen, coord.String(), face, op)
	}
}

func printD(str string, args ...any) {
	if config.Get().DevMode.Enabled {
		fmt.Printf(str+"\n", args...)
	}
}
