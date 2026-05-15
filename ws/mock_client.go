package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ognev-dev/goplease-ebitengine-client/ds"
	"github.com/ognev-dev/goplease-ebitengine-client/mock"
)

type MockClient struct {
	inbox  chan InMessage
	status ConnStatus
}

func NewMockClient() *MockClient {
	return &MockClient{
		inbox:  make(chan InMessage, 128),
		status: StatusDisconnected,
	}
}

func (m *MockClient) Inbox() <-chan InMessage {
	return m.inbox
}

func (m *MockClient) Status() ConnStatus {
	return m.status
}

func (m *MockClient) Connect(playerID string) {
	m.status = StatusConnected
	log.Printf("[mock] connected as %s", playerID)

	m.inbox <- InMessage{Action: ConnectedAction}
}

func (m *MockClient) Disconnect() {
	m.status = StatusDisconnected
}

func (m *MockClient) Send(msg OutMessage) {
	log.Printf("[mock] client sent: %s", msg.Action)

	go m.handleLogic(msg)
}

func (m *MockClient) handleLogic(msg OutMessage) {
	switch msg.Action {
	case NewGameAction:
		data, err := mock.LoadData("new_game.json")
		if err != nil {
			log.Fatal(err)
		}

		var newGameData ds.NewGamePayload
		err = json.Unmarshal(data, &newGameData)
		if err != nil {
			log.Fatal(err)
		}

		mock.NewGameState(newGameData)

		m.inbox <- InMessage{
			Action: NewGameAction,
			Data:   data,
		}

	case ReadyToPlay:
		m.inbox <- InMessage{
			Action: WaitingForOpponent,
		}

		time.Sleep(1 * time.Second)

		m.inbox <- InMessage{
			Action: PlaceUnitAction,
		}

	case EndTurnAction:
		// when player click "End Turn" | "End round", we try to determine next action
	endTurn:
		m.inbox <- InMessage{Action: WaitingForOpponent}
		time.Sleep(1 * time.Second)

		gs := mock.GetGameState()

		// if not all units played, play next
		if gs.ActiveUnit <= len(gs.UnitsQueue) && len(gs.UnitsQueue) > 0 {
			unit := gs.UnitsQueue[gs.ActiveUnit-1]
			gs.ActiveUnit++

			// for real player, just let it play
			if unit.OwnerID != mock.MockedPlayerID {
				data, err := json.Marshal(ds.PlayUnitPayload{
					UnitID: unit.ID,
				})
				if err != nil {
					log.Fatal(err)
				}
				m.inbox <- InMessage{Action: PlayUnitAction, Data: data}
				return
			}

			// for mock player simulate play
			// TODO simulate action
			m.inbox <- InMessage{Action: "unit_moved"}
			goto endTurn
			return
		}

		// check if unit can be placed on board
		if !gs.Players[0].HasPlacedUnitThisRound {
			m.inbox <- InMessage{
				Action: PlaceUnitAction,
			}

			return
		}

		if !gs.Players[1].HasPlacedUnitThisRound {
			// place random unit
			unit := mock.PickRandomUnitOfFromHandP2()
			if unit == nil {
				log.Println("[mock] place_unit: no units at hand")
				return
			}

			row, col := mock.GetRandomUnoccupiedSafeZoneCell()
			mock.PlaceUnitAt(unit, row, col)

			upl := ds.PlaceUnitPayload{
				Row:  row,
				Col:  col,
				Unit: unit,
			}

			data, err := json.Marshal(upl)
			if err != nil {
				log.Fatal(err)
			}

			gs.Players[1].HasPlacedUnitThisRound = true
			m.inbox <- InMessage{
				Action: UnitPlacedAction,
				Data:   data,
			}

			goto endTurn
		}

		// all units played if we get here, so start new round
		gs.ActiveUnit = 1
		if len(gs.Players[0].Units) > 0 {
			gs.Players[0].HasPlacedUnitThisRound = false
		}
		if len(gs.Players[1].Units) > 0 {
			gs.Players[1].HasPlacedUnitThisRound = false
		}

		goto endTurn

	case UnitPlacedAction:
		gs := mock.GetGameState()
		gs.Players[0].HasPlacedUnitThisRound = true

		data := msg.Data.(ds.UnitPlacedPayload)

		unit := mock.PickUnitFromHandByTemplateP1(data.TemplateID)
		unit.Row = data.Row
		unit.Col = data.Col

		mock.AddUnitToQueue(unit)

		// after unit placed there is nothing to do, let user end turn
		m.inbox <- InMessage{
			Action: EndRoundAction,
		}

	case CancelMatchAction:
		m.inbox <- InMessage{Action: MatchCancelledAction}
	}
}
