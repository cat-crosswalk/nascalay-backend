package ws

import (
	"errors"

	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/mitchellh/mapstructure"
)

func (c *Client) callEventHandler(req *oapi.WsJSONRequestBody) error {
	h, ok := receivedEventMap[req.Type]
	if !ok {
		return errors.New("unknown event type")
	}

	var reqBody interface{}
	if req.Body != nil {
		reqBody = *req.Body
	}

	return h(c, reqBody)
}

type eventHandler func(c *Client, body interface{}) error

var receivedEventMap = map[oapi.WsEvent]eventHandler{
	oapi.WsEventROOMSETOPTION:    (*Client).RoomSetOption,
	oapi.WsEventREQUESTGAMESTART: (*Client).RequestGameStart,
	oapi.WsEventODAIREADY:        (*Client).OdaiReady,
	oapi.WsEventODAICANCEL:       (*Client).OdaiCancel,
	oapi.WsEventODAISEND:         (*Client).OdaiSend,
	oapi.WsEventDRAWREADY:        (*Client).DrawReady,
	oapi.WsEventDRAWCANCEL:       (*Client).DrawCancel,
	oapi.WsEventDRAWSEND:         (*Client).DrawSend,
	oapi.WsEventANSWERREADY:      (*Client).AnswerReady,
	oapi.WsEventANSWERCANCEL:     (*Client).AnswerCancel,
	oapi.WsEventANSWERSEND:       (*Client).AnswerSend,
	oapi.WsEventSHOWNEXT:         (*Client).ShowNext,
	oapi.WsEventRETURNROOM:       (*Client).ReturnRoom,
}

// Receive event handlers
func (c *Client) RoomSetOption(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	_, ok := body.(oapi.WsRoomSetOptionEventBody)
	if !ok {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) RequestGameStart(_ interface{}) error {
	return nil
}

func (c *Client) OdaiReady(_ interface{}) error {
	return nil
}

func (c *Client) OdaiCancel(_ interface{}) error {
	return nil
}

func (c *Client) OdaiSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	e := new(oapi.WsOdaiSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) DrawReady(_ interface{}) error {
	return nil
}

func (c *Client) DrawCancel(_ interface{}) error {
	return nil
}

func (c *Client) DrawSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	_, ok := body.(oapi.WsDrawSendEventBody)
	if !ok {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) AnswerReady(_ interface{}) error {
	return nil
}

func (c *Client) AnswerCancel(_ interface{}) error {
	return nil
}

func (c *Client) AnswerSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	_, ok := body.(oapi.WsAnswerSendEventBody)
	if !ok {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) ShowNext(_ interface{}) error {
	return nil
}

func (c *Client) ReturnRoom(_ interface{}) error {
	return nil
}

// // Send event handlers
// func (c *Client) RoomNewMember(body interface{}) error {
// 	return nil
// }

// func (c *Client) RoomUpdateOption(body interface{}) error {
// 	return nil
// }

// func (c *Client) GameStart(body interface{}) error {
// 	return nil
// }

// func (c *Client) OdaiFinish(body interface{}) error {
// 	return nil
// }

// func (c *Client) DrawStart(body interface{}) error {
// 	return nil
// }

// func (c *Client) DrawFinish(body interface{}) error {
// 	return nil
// }

// func (c *Client) AnswerStart(body interface{}) error {
// 	return nil
// }

// func (c *Client) AnswerFinish(body interface{}) error {
// 	return nil
// }

// func (c *Client) ShowStart(body interface{}) error {
// 	return nil
// }

// func (c *Client) ShowOdai(body interface{}) error {
// 	return nil
// }

// func (c *Client) ShowCanvas(body interface{}) error {
// 	return nil
// }

// func (c *Client) ShowAnswer(body interface{}) error {
// 	return nil
// }

// func (c *Client) NextRoom(body interface{}) error {
// 	return nil
// }

// func (c *Client) ChangeHost(body interface{}) error {
// 	return nil
// }

// func (c *Client) BreakRoom(body interface{}) error {
// 	return nil
// }
