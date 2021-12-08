package ws

import (
	"encoding/json"
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
	oapi.WsEventROOMSETOPTION:    (*Client).roomSetOption,
	oapi.WsEventREQUESTGAMESTART: (*Client).requestGameStart,
	oapi.WsEventODAIREADY:        (*Client).odaiReady,
	oapi.WsEventODAICANCEL:       (*Client).odaiCancel,
	oapi.WsEventODAISEND:         (*Client).odaiSend,
	oapi.WsEventDRAWREADY:        (*Client).drawReady,
	oapi.WsEventDRAWCANCEL:       (*Client).drawCancel,
	oapi.WsEventDRAWSEND:         (*Client).drawSend,
	oapi.WsEventANSWERREADY:      (*Client).answerReady,
	oapi.WsEventANSWERCANCEL:     (*Client).answerCancel,
	oapi.WsEventANSWERSEND:       (*Client).answerSend,
	oapi.WsEventSHOWNEXT:         (*Client).showNext,
	oapi.WsEventRETURNROOM:       (*Client).returnRoom,
}

// Receive event handlers
func (c *Client) roomSetOption(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	e := new(oapi.WsRoomSetOptionEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) requestGameStart(_ interface{}) error {
	buf, err := json.Marshal(&oapi.WsGameStartEventBody{
		// TODO: 埋める
		// OdaiHint: random.OdaiExample(),
		// TimeLimit: 40,
	})
	if err != nil {
		return err
	}

	go c.sendToEachClientInRoom(buf)

	return nil
}

func (c *Client) odaiReady(_ interface{}) error {
	return nil
}

func (c *Client) odaiCancel(_ interface{}) error {
	return nil
}

func (c *Client) odaiSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	e := new(oapi.WsOdaiSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return err
	}

	return nil
}

func (c *Client) drawReady(_ interface{}) error {
	return nil
}

func (c *Client) drawCancel(_ interface{}) error {
	return nil
}

func (c *Client) drawSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	e := new(oapi.WsDrawSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) answerReady(_ interface{}) error {
	return nil
}

func (c *Client) answerCancel(_ interface{}) error {
	return nil
}

func (c *Client) answerSend(body interface{}) error {
	if body == nil {
		return errors.New("body is nil")
	}

	e := new(oapi.WsAnswerSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return errors.New("invalid body")
	}

	return nil
}

func (c *Client) showNext(_ interface{}) error {
	return nil
}

func (c *Client) returnRoom(_ interface{}) error {
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
