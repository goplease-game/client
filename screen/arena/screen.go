package arena

import (
	"encoding/json"
	"log"
	"math"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	game "github.com/ognev-dev/goplease-ebitengine-client"
	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/ui"
	"github.com/ognev-dev/goplease-ebitengine-client/ws"
)

const (
	cellSize = 64
	headerH  = 80
	statusH  = 32
	footerH  = 90
)

type Screen struct {
	server       ws.Client
	ui           *ebitenui.UI
	board        ds.Board
	roomID       string
	player       ds.Player
	opponentName string
	isMyTurn     bool

	safeZoneCells    []*DropZoneCell
	boardCellWidgets [][]*widget.Container
	unitCards        map[string]*widget.Container
	headerRef        *widget.Container
	footerRef        *widget.Container
	queuePanelRef    *widget.Container
	unitPanelRef     *widget.Container
	nextActionBtn    *widget.Button
	statusLabel      *widget.Text

	abilityPanelRef *widget.Container
	abilityPanelIn  bool

	pulseWidgets          []*widget.Container
	pulseTick             float64
	endTurnBtnPulseActive bool

	unitsQueue         []string
	activeUnitID       string
	activeUnitIndex    int
	turnNumber         int
	unitPlacedThisTurn bool
	queueIn            bool
	unitPanelIn        bool

	// ready is set to true when the server responds with phase unit_placement,
	// meaning the match has started and the local player may interact.
	ready bool
	// firstDrawn tracks whether we have completed at least one Draw call so
	// that we send ready_to_play exactly once after the UI is fully rendered.
	firstDrawn bool
}

func NewScreen(payload json.RawMessage, server ws.Client) game.Screen {
	var data ds.NewGamePayload
	if err := json.Unmarshal(payload, &data); err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}

	s := &Screen{
		server:       server,
		board:        data.Board,
		roomID:       data.RoomID,
		player:       *data.Player,
		unitCards:    make(map[string]*widget.Container),
		turnNumber:   1,
		opponentName: data.Opponent,
	}

	initDropPointAnim()
	s.setupUI(data)
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

	s.updatePulse()
	s.updateDropZoneAnim()

	s.ui.Update()
	return s, nil
}

func (s *Screen) Draw(screen *ebiten.Image) {
	s.ui.Draw(screen)

	if !s.firstDrawn {
		s.firstDrawn = true
		s.server.Send(ws.OutMessage{Action: ws.ReadyToPlay})
	}
}

// updatePulse advances the pulse animation for highlighted units and the end-turn button.
func (s *Screen) updatePulse() {
	if len(s.pulseWidgets) == 0 && !s.endTurnBtnPulseActive {
		return
	}

	s.pulseTick += 0.05
	t := (math.Sin(s.pulseTick) + 1) / 2

	if len(s.pulseWidgets) > 0 {
		c := ui.LerpColor(unitPulseColor1, unitPulseColor2, t)
		for _, w := range s.pulseWidgets {
			w.SetBackgroundImage(image.NewNineSliceColor(c))
		}
	}

	if s.endTurnBtnPulseActive && s.nextActionBtn != nil {
		s.pulseEndTurnBtn(t)
	}
}

func (s *Screen) updateDropZoneAnim() {
	animDropArrow.Update()
	for _, sc := range s.safeZoneCells {
		if sc.activeGraphic != nil {
			sc.activeGraphic.Image = animDropArrow.CurrentFrame
		}
	}
}

func (s *Screen) setupUI(data ds.NewGamePayload) {
	root := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(bodyBgColor)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	s.headerRef = s.createHeader()
	s.footerRef = s.createFooter()
	board := s.createBoardContainer(data.Board)
	statusBar := s.createStatusBar()

	root.AddChild(board)
	root.AddChild(s.headerRef)
	root.AddChild(statusBar)
	root.AddChild(s.footerRef)

	s.ui = &ebitenui.UI{Container: root}
}

func (s *Screen) setStatus(text string) {
	if s.statusLabel != nil {
		s.statusLabel.Label = text
	}
}
