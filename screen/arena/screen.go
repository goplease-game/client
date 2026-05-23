package arena

import (
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

const (
	abilityCardSize = 64 // ability card in footer panel
	unitCardSize    = 64 // unit card in hand & queue panel
	unitIconSize    = 54

	headerH = 80
	statusH = 32
	footerH = 90
)

type Screen struct {
	server ws.Client
	ui     *ebitenui.UI

	// game data
	board              ds.Board
	roomID             string
	player             ds.Player
	opponentName       string
	isMyTurn           bool
	unitsQueue         []string
	activeUnitID       string
	activeUnitIndex    int
	roundNumber        int
	unitPlacedThisTurn bool
	queueIn            bool
	unitPanelIn        bool

	safeZoneCells    []*DropZoneCell
	sortedCells      []*ui.HexCellWidget
	boardCellWidgets map[ds.HexCoord]*ui.HexCellWidget

	unitCards     map[string]*widget.Container
	headerRef     *widget.Container
	footerRef     *widget.Container
	queuePanelRef *widget.Container
	unitPanelRef  *widget.Container
	nextActionBtn *widget.Button
	statusLabel   *widget.Text

	abilityPanelRef       *widget.Container
	abilityPanelIn        bool
	abilityHighlightCells []ds.HexCoord

	pulseHexWidgets       []*ui.HexCellWidget // board hex cells that pulse
	pulseWidgets          []*widget.Container // other UI widgets that pulse
	pulseTick             float64
	endTurnBtnPulseActive bool

	// Dev panel (only active when DevMode.Enabled).
	devPanelRef       *widget.Container
	devPanelBody      *widget.Container
	devLoadList       *widget.Container
	devPanelMinimized bool

	// Movement / selection state.
	selectedUnitID  string        // unit currently selected for movement (empty = none)
	reachableCells  []ds.HexCoord // precomputed reachable positions for selectedUnit
	activeUnitMoved bool          // true once the active unit has moved this turn
	activeMoveAnim  *moveAnim     // non-nil while a movement animation is playing

	// ready is set to true when the server responds with phase unit_placement,
	// meaning the match has started and the local player may interact.
	ready bool
	// firstDrawn tracks whether we have completed at least one Draw call so
	// that we send ready_to_play exactly once after the UI is fully rendered.
	firstDrawn bool

	// for state reloading
	pendingScreen game.Screen
}

func NewScreen(snap ds.GameSnapshot, server ws.Client) game.Screen {
	s := &Screen{
		server:       server,
		board:        snap.Board,
		roomID:       snap.RoomID,
		player:       snap.Player,
		opponentName: snap.OpponentName,
		unitsQueue:   snap.UnitsQueue,
		activeUnitID: snap.ActiveUnitID,
		roundNumber:  snap.Round,
		unitCards:    make(map[string]*widget.Container),
	}

	initDropPointAnim()
	s.setupUI()
	s.restoreBoardVisuals()

	// Restore queue panel from snapshot.
	s.rebuildQueuePanel()

	return s
}
func (s *Screen) Update(g *game.Game) (game.Screen, error) {
	for {
		select {
		case msg := <-g.Server.Inbox():
			s.handleServerMessage(msg)
		default:
			goto done
		}
	}
done:

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		for coord, cell := range s.boardCellWidgets {
			if cell.HitTest(mx, my) {
				s.onCellClicked(coord)
				break
			}
		}
	}

	s.updatePulse()
	s.updateDropZoneAnim()
	s.activeMoveAnim.update()

	s.ui.Update()

	if s.pendingScreen != nil {
		return s.pendingScreen, nil
	}
	return s, nil
}

func (s *Screen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)

	if s.activeMoveAnim.active() {
		x, y := s.activeMoveAnim.currentPos()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, y)
		screen.DrawImage(s.activeMoveAnim.img, op)
	}

	if !s.firstDrawn {
		s.firstDrawn = true
		s.server.Send(ws.OutMessage{Action: ws.ReadyToPlay})
	}
}

// updatePulse advances the pulse animation for highlighted units and the end-turn button.
func (s *Screen) updatePulse() {
	if len(s.pulseHexWidgets) == 0 && len(s.pulseWidgets) == 0 && !s.endTurnBtnPulseActive {
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
}

func (s *Screen) updateDropZoneAnim() {
	animDropArrow.Update()
	for _, sc := range s.safeZoneCells {
		if sc.activeGraphic != nil {
			sc.activeGraphic = animDropArrow.CurrentFrame
		}
	}
}

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
		PostRenderHook: func(screen *ebiten.Image) {
			for _, cell := range s.boardCellWidgets {
				cell.RenderFill(screen)
			}
			s.renderGrid(screen)
			for _, cell := range s.sortedCells {
				cell.RenderUnitLayer(screen)
			}
			for _, sc := range s.safeZoneCells {
				sc.RenderAnim(screen)
			}
			for _, cell := range s.sortedCells {
				cell.RenderHUDLayer(screen)
			}
			for _, cell := range s.sortedCells {
				cell.RenderFXLayer(screen)
			}
		},
	}
}

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

func (s *Screen) setStatus(text string) {
	if s.statusLabel != nil {
		s.statusLabel.Label = text
	}
}

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

func (s *Screen) restoreSnapshot(snap ds.GameSnapshot) game.Screen {
	return NewScreen(snap, s.server)
}

func (s *Screen) restoreBoardVisuals() {
	for pos, cell := range s.board.Cells {
		if cell == nil || cell.Unit == nil {
			continue
		}

		u := *cell.Unit

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
