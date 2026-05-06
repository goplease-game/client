package client

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Layout constants ──────────────────────────────────────────────────────────

const (
	cellSize   = 50 // pixels per board cell
	boardCols  = 12
	boardRows  = 12
	boardOffX  = (ScreenWidth - cellSize*boardCols) / 2
	boardOffY  = 20
	handPanelY = boardOffY + cellSize*boardRows + 8
)

// ── Lightweight state types (mirrors server snapshots) ───────────────────────

type UnitSnap struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	OwnerID    string `json:"owner_id"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"max_hp"`
	Col        int    `json:"col"`
	Row        int    `json:"row"`
	Upgraded   bool   `json:"upgraded"`
}

type StateSnap struct {
	Phase        string `json:"phase"`
	CurrentTurn  int    `json:"current_turn"`
	MaxTurns     int    `json:"max_turns"`
	ActivePlayer int    `json:"active_player"`
	Board        struct {
		Units []UnitSnap `json:"units"`
	} `json:"board"`
	Players []struct {
		ID       string     `json:"id"`
		Name     string     `json:"name"`
		HandSize int        `json:"hand_size"`
		Hand     []UnitSnap `json:"hand,omitempty"`
	} `json:"players"`
}

// ── RoomScreen ────────────────────────────────────────────────────────────────

type RoomScreen struct {
	roomID string
	state  StateSnap

	// Placement interaction
	selectedHandIdx int // index into my hand, -1 = none
	hoveredCol      int
	hoveredRow      int

	statusLine string
}

func NewRoomScreen(roomID string, initialState json.RawMessage) *RoomScreen {
	s := &RoomScreen{
		roomID:          roomID,
		selectedHandIdx: -1,
		statusLine:      "Place your units, then press End Turn",
	}
	_ = json.Unmarshal(initialState, &s.state)
	return s
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *RoomScreen) Update(g *Game) (Screen, error) {
	// Drain inbox
	for {
		select {
		case msg := <-g.Server.Inbox:
			s.handleMessage(g, msg)
		default:
			goto doneInbox
		}
	}
doneInbox:

	s.updateMouse(g)
	s.updateKeys(g)
	return s, nil
}

func (s *RoomScreen) handleMessage(g *Game, msg WSMessage) {
	switch msg.Action {
	case "state_update", "unit_placed", "unit_recalled":
		_ = json.Unmarshal(msg.Data, &s.state)

	case "turn_result":
		var payload struct {
			NewPhase string `json:"new_phase"`
		}
		_ = json.Unmarshal(msg.Data, &payload)
		s.statusLine = fmt.Sprintf("Simulation done. Phase: %s", payload.NewPhase)
		// Also expect a follow-up state_update from the server.

	case "game_over":
		var payload struct {
			Winner string `json:"winner"`
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(msg.Data, &payload)
		if payload.Winner == g.PlayerID {
			s.statusLine = "You WIN! (" + payload.Reason + ")"
		} else {
			s.statusLine = "You lose. (" + payload.Reason + ")"
		}

	case "error":
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(msg.Data, &e)
		s.statusLine = "Error: " + e.Message
	}
}

func (s *RoomScreen) updateMouse(g *Game) {
	mx, my := ebiten.CursorPosition()

	// Board hover
	col := (mx - boardOffX) / cellSize
	row := (my - boardOffY) / cellSize
	if col >= 0 && col < boardCols && row >= 0 && row < boardRows {
		s.hoveredCol = col
		s.hoveredRow = row
	} else {
		s.hoveredCol = -1
		s.hoveredRow = -1
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if s.selectedHandIdx >= 0 && s.hoveredCol >= 0 {
			// Place selected unit
			myHand := s.myHand(g.PlayerID)
			if s.selectedHandIdx < len(myHand) {
				unit := myHand[s.selectedHandIdx]
				g.Server.Send(map[string]any{
					"type":             "place_unit",
					"unit_instance_id": unit.InstanceID,
					"col":              s.hoveredCol,
					"row":              s.hoveredRow,
				})
				s.selectedHandIdx = -1
			}
		} else if s.hoveredCol >= 0 {
			// Recall unit from board
			g.Server.Send(map[string]any{
				"type": "recall_unit",
				"col":  s.hoveredCol,
				"row":  s.hoveredRow,
			})
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		s.selectedHandIdx = -1
	}

	// Hand slot clicks
	myHand := s.myHand(g.PlayerID)
	for i := range myHand {
		sx := boardOffX + i*52
		sy := handPanelY
		if mx >= sx && mx < sx+48 && my >= sy && my < sy+48 {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				s.selectedHandIdx = i
			}
		}
	}
}

func (s *RoomScreen) updateKeys(g *Game) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Server.Send(map[string]any{"type": "end_turn"})
		s.statusLine = "Waiting for simulation…"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.selectedHandIdx = -1
	}
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (s *RoomScreen) Draw(screen *ebiten.Image) {
	s.drawBoard(screen)
	s.drawHand(screen)
	s.drawHUD(screen)
}

func (s *RoomScreen) drawBoard(screen *ebiten.Image) {
	for row := 0; row < boardCols; row++ {
		for col := 0; col < boardRows; col++ {
			x := float32(boardOffX + col*cellSize)
			y := float32(boardOffY + row*cellSize)

			// Cell background
			bg := color.NRGBA{R: 30, G: 30, B: 45, A: 255}
			if row < 2 {
				bg = color.NRGBA{R: 20, G: 50, B: 80, A: 255} // player 1 safe zone
			} else if row >= boardCols-2 {
				bg = color.NRGBA{R: 80, G: 20, B: 20, A: 255} // player 2 safe zone
			}
			if col == s.hoveredCol && row == s.hoveredRow {
				bg = color.NRGBA{R: 60, G: 60, B: 90, A: 255}
			}
			vector.DrawFilledRect(screen, x, y, cellSize-1, cellSize-1, bg, false)

			// Grid outline
			vector.StrokeRect(screen, x, y, cellSize-1, cellSize-1,
				1, color.NRGBA{R: 50, G: 50, B: 70, A: 255}, false)
		}
	}

	// Draw units on board
	for _, u := range s.state.Board.Units {
		s.drawUnit(screen, u,
			boardOffX+u.Col*cellSize,
			boardOffY+u.Row*cellSize)
	}
}

func (s *RoomScreen) drawUnit(screen *ebiten.Image, u UnitSnap, px, py int) {
	// Unit background colour by owner (player 1 = blue, player 2 = red)
	c := color.NRGBA{R: 60, G: 100, B: 200, A: 220}
	if len(s.state.Players) > 1 && u.OwnerID == s.state.Players[1].ID {
		c = color.NRGBA{R: 200, G: 60, B: 60, A: 220}
	}
	if u.Upgraded {
		c.A = 255
		c.R += 30
	}
	vector.DrawFilledRect(screen, float32(px+2), float32(py+2), cellSize-5, cellSize-5, c, false)

	// HP bar
	barW := float32(cellSize - 6)
	hpFrac := float32(u.HP) / float32(u.MaxHP)
	vector.DrawFilledRect(screen, float32(px+3), float32(py+cellSize-8), barW*hpFrac, 4,
		color.NRGBA{R: 80, G: 220, B: 80, A: 255}, false)

	// Name abbreviation (first 3 chars)
	label := u.Name
	if len(label) > 3 {
		label = label[:3]
	}
	ebitenutil.DebugPrintAt(screen, label, px+4, py+4)
}

func (s *RoomScreen) drawHand(screen *ebiten.Image) {
	myHand := s.myHand("")
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Hand (%d):", len(myHand)), boardOffX, handPanelY-14)

	for i, u := range myHand {
		x := boardOffX + i*52
		y := handPanelY

		bg := color.NRGBA{R: 40, G: 80, B: 140, A: 255}
		if i == s.selectedHandIdx {
			bg = color.NRGBA{R: 80, G: 160, B: 255, A: 255}
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), 48, 48, bg, false)
		label := u.Name
		if len(label) > 3 {
			label = label[:3]
		}
		ebitenutil.DebugPrintAt(screen, label, x+4, y+4)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", u.HP, u.MaxHP), x+2, y+30)
	}
}

func (s *RoomScreen) drawHUD(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Turn %d/%d  Phase: %s",
			s.state.CurrentTurn, s.state.MaxTurns, s.state.Phase),
		4, 4)
	ebitenutil.DebugPrintAt(screen, s.statusLine, 4, ScreenHeight-20)
	ebitenutil.DebugPrintAt(screen, "[Enter] End Turn   [Esc] Deselect", 4, ScreenHeight-36)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// myHand returns the local player's hand units from the last state snapshot.
// If playerID is empty it returns the first player's hand (for drawing the hand
// during development before the player ID is wired up).
func (s *RoomScreen) myHand(playerID string) []UnitSnap {
	for _, p := range s.state.Players {
		if playerID == "" || p.ID == playerID {
			return p.Hand
		}
	}
	return nil
}
