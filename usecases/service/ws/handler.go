package ws

import (
	"encoding/json"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/mitchellh/mapstructure"
)

func (c *Client) callEventHandler(req *oapi.WsJSONRequestBody) error {
	h, ok := receivedEventMap[req.Type]
	if !ok {
		return errUnknownEventType
	}

	return h(c, req.Body)
}

type eventHandler func(c *Client, body interface{}) error

var receivedEventMap = map[oapi.WsEvent]eventHandler{
	oapi.WsEventROOMSETOPTION:    (*Client).receiveRoomSetOptionEvent,
	oapi.WsEventREQUESTGAMESTART: (*Client).receiveRequestGameStartEvent,
	oapi.WsEventODAIREADY:        (*Client).receiveOdaiReadyEvent,
	oapi.WsEventODAICANCEL:       (*Client).receiveOdaiCancelEvent,
	oapi.WsEventODAISEND:         (*Client).receiveOdaiSendEvent,
	oapi.WsEventDRAWREADY:        (*Client).receiveDrawReadyEvent,
	oapi.WsEventDRAWCANCEL:       (*Client).receiveDrawCancelEvent,
	oapi.WsEventDRAWSEND:         (*Client).receiveDrawSendEvent,
	oapi.WsEventANSWERREADY:      (*Client).receiveAnswerReadyEvent,
	oapi.WsEventANSWERCANCEL:     (*Client).receiveAnswerCancelEvent,
	oapi.WsEventANSWERSEND:       (*Client).receiveAnswerSendEvent,
	oapi.WsEventSHOWNEXT:         (*Client).receiveShowNextEvent,
	oapi.WsEventRETURNROOM:       (*Client).receiveReturnRoomEvent,
}

// Receive event handlers
func (c *Client) receiveRoomSetOptionEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsRoomSetOptionEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) receiveRequestGameStartEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	if c.userId != c.room.HostId {
		return errUnAuthorized
	}

	c.room.Game.Status = model.GameStatusOdai

	if err := c.sendGameStartEvent(); err != nil {
		return err
	}

	return nil
}

func (c *Client) receiveOdaiReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.room.Game.AddReady(c.userId)

	if c.room.AllMembersAreReady() {
		if err := c.sendOdaiFinishEvent(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) receiveOdaiCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.room.Game.CancelReady(c.userId)

	return nil
}

func (c *Client) receiveOdaiSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsOdaiSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) receiveDrawReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) receiveDrawCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) receiveDrawSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsDrawSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) receiveAnswerReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) receiveAnswerCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) receiveAnswerSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsAnswerSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) receiveShowNextEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) receiveReturnRoomEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// // Send event handlers
// func (c *Client) sendRoomNewMemberEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusRoom) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

func (c *Client) sendRoomUpdateOptionEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	return nil
}

func (c *Client) sendGameStartEvent() error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventGAMESTART,
			Body: &oapi.WsGameStartEventBody{
				// TODO: 埋める
				// OdaiHint: random.OdaiExample(),
				// TimeLimit: 40,
			},
		},
	)
	if err != nil {
		return err
	}

	go c.sendToEachClientInRoom(buf)

	return nil
}

func (c *Client) sendOdaiFinishEvent() error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventODAIFINISH,
		},
	)
	if err != nil {
		return err
	}

	go c.sendToEachClientInRoom(buf)

	return nil
}

// func (c *Client) sendDrawStartEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusDraw) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendDrawFinishEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusDraw) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendAnswerStartEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusAnswer) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendAnswerFinishEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusAnswer) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendShowStartEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusShow) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendShowOdaiEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusShow) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendShowCanvasEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusShow) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendShowAnswerEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusShow) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendNextRoomEvent(body interface{}) error {
// 	if !c.room.GameStatusIs(model.GameStatusShow) {
// 		return errWrongPhase
// 	}

// 	return nil
// }

// func (c *Client) sendChangeHostEvent(body interface{}) error {
// 	return nil
// }

// func (c *Client) sendBreakRoomEvent(body interface{}) error {
// 	return nil
// }
