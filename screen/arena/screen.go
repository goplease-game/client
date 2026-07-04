package arena

import (
	"fmt"
	"image/color"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	game "github.com/goplease-game/client"
	"github.com/goplease-game/client/config"
	"github.com/goplease-game/client/ds"
	"github.com/goplease-game/client/ui"
	"github.com/goplease-game/client/ws"
	"github.com/goplease-game/server/ability"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
)

// Screen is the main arena game screen.
// It owns all game state visible to the local player and orchestrates
// server communication, UI layout, board rendering, and animations.
type Screen struct {
	snapshot   ds.GameSnapshot
	server     ws.Client
	ui         *ebitenui.UI
	menuUI     *ebitenui.UI
	gameOverUI *ebitenui.UI

	tutorialUI      *ebitenui.UI
	tutorialOverlay *TutorialOverlay

	// Game state received from the server.
	roomID             string // todo: arenaID
	board              ds.Board
	player             ds.Player
	opponentName       string
	isMyTurn           bool
	unitsQueue         []*ds.Unit
	activeUnitID       string
	prevActiveUnitID   string
	roundNumber        int
	unitPlacedThisTurn bool
	queueIn            bool
	unitPanelIn        bool

	// Board rendering.
	dropZoneCells    []*DropZoneCell     // safe-zone cells that accept unit drops
	sortedCells      []*ui.HexCellWidget // board cells sorted by (R, Q) for deterministic overlay render order
	boardCellWidgets map[ds.HexCoord]*ui.HexCellWidget

	// UI widget references used for dynamic updates.
	unitCards     map[string]*widget.Container // unitID:widget
	headerRef     *widget.Container
	footerRef     *widget.Container
	queuePanelRef *widget.Container
	unitPanelRef  *widget.Container
	statusBarRef  *widget.Container
	nextActionBtn *widget.Button
	statusLabel   *widget.Text
	statusText    string

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

	// you can see hex coords by holding Alt
	showCellCoordinates bool

	// Movement and selection state.
	selectedUnitID    string        // unit currently selected for movement; empty means none
	reachableCells    []ds.HexCoord // precomputed reachable positions for selectedUnitID
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

	// nextScreen is set when a screen transition should occur on the next Update.
	nextScreen game.Screen

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

	menuOverlayRef  *widget.Container
	menuVisible     bool
	gameOverVisible bool

	maxPhantomAPPerUnitPerTurn int

	OnExitScreen    func() game.Screen
	OnRestartScreen func() game.Screen

	nextActionHourglass *widget.Graphic

	logWindow         *gameLogWindow
	logPanelRef       *widget.Container
	boardContainerRef *widget.Container

	leftPanelRef   *widget.Container
	infoPanelRef   *widget.Container
	infoPanelUnit  *ds.Unit
	infoPanelDirty bool

	abilitySlots []abilitySlot
}

// NewScreen constructs a fully initialised arena Screen from a server snapshot.
func NewScreen(snap ds.GameSnapshot, server ws.Client) *Screen {
	s := &Screen{
		snapshot:        snap,
		server:          server,
		board:           snap.Board,
		roomID:          snap.ArenaID,
		player:          snap.Player,
		opponentName:    snap.OpponentName,
		unitsQueue:      snap.UnitsQueue,
		activeUnitID:    snap.ActiveUnitID,
		roundNumber:     snap.Round,
		unitCards:       make(map[string]*widget.Container),
		turnTimeSeconds: snap.TurnTimeSeconds,

		maxPhantomAPPerUnitPerTurn: snap.MaxPhantomAPPerUnitPerTurn,
	}

	initDropPointAnim()
	s.setupUI()
	s.setupTutorial()
	s.restoreBoardVisuals()
	s.rebuildQueuePanel()

	return s
}

// Update processes server messages, handles input, and advances all animations.
// Implements game.Screen.
func (s *Screen) Update(_ *game.Game) (game.Screen, error) {
	keys := config.Get().Keybindings
	// Drain all pending server messages before updating game logic.
	for {
		select {
		case msg := <-s.server.Inbox():
			s.handleServerMessage(msg)
		default:
			goto done
		}
	}
done:

	tutorialWasVisible := s.tutorialOverlay != nil && s.tutorialOverlay.IsVisible()

	s.updateDelayedActions()
	s.updateNewRoundBanner()
	s.updateFloatingTexts()

	for slot, as := range s.abilitySlots {
		if game.KeyJustPressed(abilityKeyForSlot(slot)) {
			s.activateAbilitySlot(as.ability, as.card, as.bgColor, as.iconGraphic)
			break
		}
	}

	// Handle hex cell clicks manually since HexCellWidget uses custom hit testing.
	if !tutorialWasVisible && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		hit := false
		for coord, cell := range s.boardCellWidgets {
			if cell.HitTest(mx, my) {
				s.onCellClicked(coord)
				hit = true
				break
			}
		}
		// If the click missed every board cell (e.g. landed on empty board space)
		// while a unit was selected for movement, cancel the selection — clicking
		// anywhere outside the reachable cells should deselect, not just clicking
		// another cell.
		if !hit && s.selectedUnitID != "" {
			u := s.unitByID(s.selectedUnitID)
			s.deselectUnit()
			s.showAbilityPanel(u)
		}
	}

	if !tutorialWasVisible && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if s.selectedAbility != nil {
			s.cancelAbilitySelection()
		} else {
			s.toggleGameMenu()
		}
	}

	if game.KeyJustPressed(keys.ShowGameLog) {
		s.toggleGameLog()
	}

	if game.KeyJustPressed(keys.EndTurn) {
		s.endTurn()
	}

	if !tutorialWasVisible && game.KeyJustPressed(keys.Move) {
		if u := s.unitByID(s.activeUnitID); u != nil {
			s.onMoveButtonClicked(u)
		}
	}

	if game.KeyJustPressed(keys.ShowCoordinates) {
		s.showCellCoordinates = game.KeyPressed(keys.ShowCoordinates)
	}

	s.updatePulse()
	s.updateDropZoneAnim()
	s.updateMoveAnimations()
	s.updateFxAnims()

	if s.tutorialOverlay != nil && s.tutorialOverlay.IsVisible() {
		s.tutorialOverlay.UI.Update()
	}

	if !tutorialWasVisible {
		if s.gameOverVisible {
			s.gameOverUI.Update()
		} else if s.menuVisible {
			s.menuUI.Update()
		}
		s.ui.Update()
	}

	if s.infoPanelUnit != nil && s.infoPanelDirty {
		s.showInfoPanel(s.buildUnitInfoPanel(s.infoPanelUnit))
		s.infoPanelDirty = false
	}

	if s.nextScreen != nil {
		return s.nextScreen, nil
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

	if u := s.unitByID(s.activeUnitID); u != nil {
		if cell, ok := s.boardCellWidgets[u.Pos]; ok {
			s.drawAPMarkers(screen, u, cell)
		}
	}

	if s.tutorialOverlay != nil && s.tutorialOverlay.IsVisible() {
		s.tutorialOverlay.UI.Draw(screen)
	}

	if s.menuVisible {
		s.menuUI.Draw(screen)
	}
	if s.gameOverVisible {
		s.gameOverUI.Draw(screen)
	}

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

		keys := config.Get().Keybindings
		if keys.ShowGameLog != nil {
			s.appendLogEntry(logMessage{
				Text: fmt.Sprintf("Press <ability>[%s]</ability> to toggle this log", game.KeyName(keys.ShowGameLog)),
			})
		}
		if keys.ShowCoordinates != nil {
			s.appendLogEntry(logMessage{
				Text: fmt.Sprintf("Hold <ability>[%s]</ability> to show cell coordinates", game.KeyName(keys.ShowCoordinates)),
			})
		}

		s.appendLogEntry(logMessage{Text: ""})
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
	for _, sc := range s.dropZoneCells {
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
	s.leftPanelRef = s.createLeftPanel()
	board := s.createBoardContainer()
	statusBar := s.createStatusBar()

	root.AddChild(board)
	root.AddChild(s.headerRef)
	root.AddChild(statusBar)
	root.AddChild(s.footerRef)
	root.AddChild(s.leftPanelRef)

	s.setupDevPanel(root)

	if config.Get().ShowGameLog {
		s.toggleGameLog()
	}

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
			for _, sc := range s.dropZoneCells {
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

			if s.showCellCoordinates {
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

	s.menuUI = &ebitenui.UI{Container: s.menuOverlayRef}
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
	s.statusText = text
	s.refreshStatusBar()
}

func (s *Screen) refreshStatusBar() {
	if s.statusBarRef == nil {
		return
	}

	text := s.statusText
	if s.tutorialOverlay != nil && s.tutorialOverlay.IsVisible() {
		text = ""
	}

	// Clear previous text widget to reset cached MinSize dimensions.
	s.statusBarRef.RemoveChildren()

	tf := ui.TextFace(16)
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

	canAct := s.unitCanAct(u)

	var moveKey string
	if key := config.Get().Keybindings.Move; key != nil {
		moveKey = " [" + game.KeyName(key) + "] "
	}

	var status string
	switch {
	case u.CanMove() && canAct:
		status = u.Name + " can move " + moveKey + " and use an ability"
	case u.CanMove():
		status = u.Name + " can move" + moveKey
	case canAct:
		status = u.Name + " can use an ability"
	default:
		status = u.Name + " may end turn"
		if key := config.Get().Keybindings.EndTurn; key != nil {
			status += " [" + game.KeyName(key) + "]"
		}
	}

	s.setStatus(status)
}

// takeSnapshot captures the current game state into a GameSnapshot.
// Used before navigating away from the screen so state can be restored.
func (s *Screen) takeSnapshot() ds.GameSnapshot {
	return ds.GameSnapshot{
		ArenaID:      s.roomID,
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
		s.buildBoardCard(w, u)
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

	canAct := s.unitCanAct(u)

	if !u.CanMove() && !canAct {
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

// this is what laziness has done to me.
func printD(str string, args ...any) {
	if config.Get().DevMode.Enabled {
		fmt.Printf(str+"\n", args...)
	}
}
