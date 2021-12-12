package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/util/canvas"
	"github.com/21hack02win/nascalay-backend/util/random"
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

// ROOM_NEW_MEMBER
// 部屋に追加のメンバーが来たことを通知する (サーバー -> ルーム全員)
func (c *Client) sendRoomNewMemberEvent(room *model.Room) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventROOMNEWMEMBER,
			Body: oapi.WsRoomNewMemberEventBody{
				Capacity: room.Capacity.Int(),
				HostId:   room.HostId.UUID(),
				Members:  oapi.RefillUsers(room.Members),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// ROOM_SET_OPTION
// ゲームのオプションを設定する (ホスト -> サーバー)
func (c *Client) receiveRoomSetOptionEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsRoomSetOptionEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	updateBody := new(oapi.WsRoomUpdateOptionEventBody)
	game := c.room.Game

	// Set options
	if e.TimeLimit != nil {
		game.TimeLimit = model.TimeLimit(*e.TimeLimit)
		updateBody.TimeLimit = e.TimeLimit
	}

	if err := c.sendRoomUpdateOptionEvent(updateBody); err != nil {
		return fmt.Errorf("failed to send ROOM_UPDATE_OPTION event: %w", err)
	}

	return nil
}

// ROOM_UPDATE_OPTION
// ゲームの設定を更新する (サーバー -> ルーム全員)
func (c *Client) sendRoomUpdateOptionEvent(body *oapi.WsRoomUpdateOptionEventBody) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventROOMUPDATEOPTION,
			Body: body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// REQUEST_GAME_START
// ゲームを開始する (ホスト -> サーバー)
func (c *Client) receiveRequestGameStartEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}
	if len(c.room.Members) < 2 {
		return errNotEnoughMember
	}

	if stopped := c.room.Game.BreakTimer.Stop(); !stopped {
		go c.waitAndBreakRoom()
	}

	if c.userId != c.room.HostId {
		return errUnAuthorized
	}

	if err := c.sendGameStartEvent(); err != nil {
		return fmt.Errorf("failed to send GAME_START event: %w", err)
	}

	return nil
}

// GAME_START
// ゲームの開始を通知する (サーバー -> ルーム全員)
// ODAIフェーズを開始する
func (c *Client) sendGameStartEvent() error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	for _, m := range c.room.Members {
		buf, err := json.Marshal(
			&oapi.WsJSONBody{
				Type: oapi.WsEventGAMESTART,
				Body: &oapi.WsGameStartEventBody{
					OdaiExample: random.OdaiExample(),
					TimeLimit:   int(c.room.Game.TimeLimit),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to encode as JSON: %w", err)
		}

		cc, ok := c.hub.userIdToClient[m.Id]
		if !ok {
			return fmt.Errorf("failed to find client: %w", errNotFound)
		}

		cc.sendMsg(buf)
	}

	// ODAIフェーズに移行
	c.room.Game.Status = model.GameStatusOdai
	// ODAIのカウントダウン開始
	if !c.room.Game.Timer.Stop() {
		<-c.room.Game.Timer.C
	}
	c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
	c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

	go func() {
		select {
		case <-c.room.Game.Timer.C:
			if err := c.sendOdaiFinishEvent(); err != nil {
				log.Println("failed to send ODAI_FINISH event: ", err.Error())
			}
		case <-c.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ODAI_READY
// お題の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveOdaiReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.allMembersAreReady() {
		if c.room.Game.Timer.Stop() {
			c.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.room.Game.Timer.C
		}
		if err := c.sendOdaiFinishEvent(); err != nil {
			return fmt.Errorf("failed to send ODAI_FINISH event: %w", err)
		}
	}

	return nil
}

// ODAI_CANCEL
// お題の入力の完了を解除する (ルームの各員 -> サーバー)
func (c *Client) receiveOdaiCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	return nil
}

// ODAI_FINISH
// 全員がお題の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントはお題を送信する
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
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// ODAI_SEND
// お題を送信する (ルームの各員 -> サーバー)
// DRAWフェーズを開始する
func (c *Client) receiveOdaiSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsOdaiSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	// 存在チェック
	for _, v := range c.room.Game.Odais {
		if v.SenderId == c.userId || v.Title == model.OdaiTitle(e.Odai) {
			return errAlreadyExists
		}
	}

	c.room.Game.AddOdai(c.userId, model.OdaiTitle(e.Odai))

	// 全員のお題送信が完了したらDRAWフェーズに移行
	members := c.room.Members
	game := c.room.Game
	if len(game.Odais) == len(members) {
		game.ResetReady()
		game.Status = model.GameStatusDraw
		game.DrawCount = 0
		random.SetupMemberRoles(game, c.room.Members)

		if err := c.sendDrawStartEvent(); err != nil {
			return fmt.Errorf("failed to send DRAW_START event: %w", err)
		}

		// DRAWのカウントダウン開始
		c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
		c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

		go func() {
			select {
			case <-c.room.Game.Timer.C:
				if err := c.sendDrawFinishEvent(); err != nil {
					log.Println("failed to send DRAW_FINISH event: ", err.Error())
				}
			case <-c.room.Game.TimerStopChan:
			}
		}()
	}

	return nil
}

// DRAW_START
// キャンバス情報とお題を送信する (サーバー -> ルーム各員)
func (c *Client) sendDrawStartEvent() error {
	// NOTE: 最後にお題を送信したユーザー(2回目以降は最後に絵を送信したユーザー)のクライアントから実行される
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	var (
		game      = c.room.Game
		drawCount = game.DrawCount
	)

	for _, o := range game.Odais {
		if len(o.DrawerSeq) <= drawCount.Int() {
			return errInvalidDrawCount
		}

		drawer := o.DrawerSeq[drawCount.Int()]
		cc, ok := c.hub.userIdToClient[drawer.UserId] // `drawCount`番目に描くユーザーのクライアント
		if !ok {
			return errNotFound
		}

		drawnArea := make([]int, drawCount.Int())
		for i := 0; i < drawCount.Int(); i++ {
			drawnArea[i] = o.DrawerSeq[i].AreaId.Int()
		}

		buf, err := json.Marshal(
			&oapi.WsJSONBody{
				Type: oapi.WsEventDRAWSTART,
				Body: oapi.WsDrawStartEventBody{
					AllDrawPhaseNum: c.room.AllDrawPhase(),
					Canvas: oapi.Canvas{
						AreaId:    drawer.AreaId.Int(),
						BoardName: game.Canvas.BoardName,
					},
					DrawPhaseNum: game.DrawCount.Int(),
					Img:          o.Img.AddPrefix(),
					Odai:         o.Title.String(),
					TimeLimit:    int(game.TimeLimit),
					DrawnArea:    drawnArea,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to encode as JSON: %w", err)
		}

		cc.sendMsg(buf)
	}

	return nil
}

// DRAW_READY
// 絵が書き終わっていることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveDrawReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.allMembersAreReady() {
		if c.room.Game.Timer.Stop() {
			c.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.room.Game.Timer.C
		}
		if err := c.sendDrawFinishEvent(); err != nil {
			return fmt.Errorf("failed to send DRAW_FINISH event: %w", err)
		}
	}

	return nil
}

// DRAW_CANCEL
// 絵が書き終わっている通知を解除する (ルームの各員 -> サーバー)
func (c *Client) receiveDrawCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	return nil
}

// DRAW_FINISH
// 全員が絵を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは絵を送信する
func (c *Client) sendDrawFinishEvent() error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventDRAWFINISH,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// DRAW_SEND
// 絵を送信する (ルームの各員 -> サーバー)
// お題が残っていたら再度DRAW_START が送信される
// お題がすべて終わったらANSWERフェーズを開始する
func (c *Client) receiveDrawSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsDrawSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	for _, v := range c.room.Game.Odais {
		if v.DrawerSeq[c.room.Game.DrawCount].UserId == c.userId {
			sendImg := e.Img[strings.IndexByte(e.Img, ',')+1:]
			if len(v.Img) > 0 {
				newImg, err := canvas.MergeImage(string(v.Img), sendImg)
				if err != nil {
					v.Img = model.Img(sendImg)
					return fmt.Errorf("failed to merge image: %w", err)
				}
				v.Img = model.Img(newImg)
			} else {
				v.Img = model.Img(sendImg)
			}
			v.ImgUpdated = true
			break
		}
	}

	// 全員の絵の送信が完了したら再度DRAW_STARTを送信する or ANSWERフェーズに移行
	allImgUpdated := true
	for _, v := range c.room.Game.Odais {
		if !v.ImgUpdated {
			allImgUpdated = false
			break
		}
	}

	game := c.room.Game
	if allImgUpdated {
		game.ResetReady()
		if game.DrawCount.Int()+1 < c.room.AllDrawPhase() {
			game.DrawCount++

			game.ResetImgUpdated()

			if err := c.sendDrawStartEvent(); err != nil {
				return fmt.Errorf("failed to send DRAW_START event: %w", err)
			}

			// DRAWのカウントダウン開始
			c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
			c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

			go func() {
				select {
				case <-c.room.Game.Timer.C:
					if err := c.sendDrawFinishEvent(); err != nil {
						log.Println("failed to send DRAW_FINISH event: ", err.Error())
					}
				case <-c.room.Game.TimerStopChan:
				}
			}()
		} else {
			game.Status = model.GameStatusAnswer

			if err := c.sendAnswerStartEvent(); err != nil {
				return fmt.Errorf("failed to send ANSWER_START event: %w", err)
			}
		}
	}

	return nil
}

// ANSWER_START
// 絵が飛んできて，回答する (サーバー -> ルーム各員)
func (c *Client) sendAnswerStartEvent() error {
	// NOTE: 最後にお題を送信した人のクライアントで行う
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	for _, v := range c.room.Game.Odais {
		ac, ok := c.hub.userIdToClient[v.AnswererId]
		if !ok {
			return errNotFound
		}

		buf, err := json.Marshal(
			&oapi.WsJSONBody{
				Type: oapi.WsEventANSWERSTART,
				Body: oapi.WsAnswerStartEventBody{
					Img:       v.Img.AddPrefix(),
					TimeLimit: int(c.room.Game.TimeLimit),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to encode as JSON: %w", err)
		}

		ac.sendMsg(buf)
	}

	// ANSWERのカウントダウン開始
	c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
	c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

	go func() {
		select {
		case <-c.room.Game.Timer.C:
			if err := c.sendAnswerFinishEvent(); err != nil {
				log.Println("failed to send ANSWER_FINISH event: ", err.Error())
			}
		case <-c.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ANSWER_READY
// 回答の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveAnswerReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.allMembersAreReady() {
		if c.room.Game.Timer.Stop() {
			c.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.room.Game.Timer.C
		}
		if err := c.sendAnswerFinishEvent(); err != nil {
			return fmt.Errorf("failed to send ANSWER_FINISH event: %w", err)
		}
	}

	return nil
}

// ANSWER_CANCEL
// 回答の入力の完了を解除する (ルームの各員 -> サーバー)
func (c *Client) receiveAnswerCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	return nil
}

// ANSWER_FINISH
// 全員が回答の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは回答を送信する
func (c *Client) sendAnswerFinishEvent() error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventANSWERFINISH,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// ANSWER_SEND
// 回答を送信する (ルームの各員 -> サーバー)
// SHOWフェーズを開始する
func (c *Client) receiveAnswerSendEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsAnswerSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	game := c.room.Game

	for _, v := range game.Odais {
		if v.AnswererId == c.userId {
			ma := model.OdaiAnswer(e.Answer)
			v.Answer = &ma
			break
		}
	}

	// 全員の回答が送信されたらSHOWフェーズに移行
	allAnswerReceived := true
	for _, v := range game.Odais {
		if v.Answer == nil {
			allAnswerReceived = false
			break
		}
	}

	if allAnswerReceived {
		game.Status = model.GameStatusShow

		if err := c.sendShowStartEvent(); err != nil {
			return fmt.Errorf("failed to send SHOW_START event: %w", err)
		}
	}

	return nil
}

// SHOW_START
// 結果表示フェーズが始まったことを通知する (サーバー -> ルーム全員)
func (c *Client) sendShowStartEvent() error {
	// NOTE: 最後に回答したユーザーのクライアントで行う
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	c.resetBreakTimer()

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWSTART,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsgToEachClientInRoom(buf)

	return nil
}

// SHOW_NEXT
// つぎの結果表示を要求する (ホスト -> サーバー)
func (c *Client) receiveShowNextEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	if c.userId != c.room.HostId {
		return errUnAuthorized
	}

	switch c.room.Game.NextShowPhase {
	case model.GameShowPhaseOdai:
		c.bloadcast(func(cc *Client) {
			if err := cc.sendShowOdaiEvent(); err != nil {
				log.Println("failed to send SHOW_ODAI event:", err.Error())
			}
		})
	case model.GameShowPhaseCanvas:
		c.bloadcast(func(cc *Client) {
			if err := cc.sendShowCanvasEvent(); err != nil {
				log.Println("failed to send SHOW_CANVAS event:", err.Error())
			}
		})
	case model.GameShowPhaseAnswer:
		c.bloadcast(func(cc *Client) {
			if err := cc.sendShowAnswerEvent(); err != nil {
				log.Println("failed to send SHOW_ANSWER event:", err.Error())
			}
		})
	case model.GameShowPhaseEnd:
	default:
		return errUnknownPhase
	}

	return nil
}

// SHOW_ODAI
// 最初のお題を受信する (サーバー -> ルーム全員)
func (c *Client) sendShowOdaiEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}
	var (
		game = c.room.Game
		sc   = game.ShowCount.Int()
	)
	if len(game.Odais) < sc+1 {
		return errNotFound
	}

	sender := oapi.User{}
	for _, m := range c.room.Members {
		if m.Id == game.Odais[sc].SenderId {
			sender = oapi.RefillUser(&m)
			break
		}
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWODAI,
			Body: &oapi.WsShowOdaiEventBody{
				Sender: sender,
				Next:   oapi.WsNextShowStatus("canvas"),
				Odai:   game.Odais[sc].Title.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	if c.userId == c.room.HostId {
		c.room.Game.NextShowPhase = model.GameShowPhaseCanvas
	}

	c.sendMsg(buf)

	return nil
}

// SHOW_CANVAS
// 次のキャンバスを受信する (サーバー -> ルーム全員)
func (c *Client) sendShowCanvasEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	var (
		game = c.room.Game
		sc   = game.ShowCount.Int()
	)

	if len(game.Odais) < sc+1 {
		return errNotFound
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWCANVAS,
			Body: &oapi.WsShowCanvasEventBody{
				Next: oapi.WsNextShowStatus("answer"),
				Img:  game.Odais[sc].Img.AddPrefix(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	if c.userId == c.room.HostId {
		c.room.Game.NextShowPhase = model.GameShowPhaseAnswer
	}

	c.sendMsg(buf)

	return nil
}

// SHOW_ANSWER
// 最後の回答を受信する (サーバー -> ルーム全員)
func (c *Client) sendShowAnswerEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	var (
		game = c.room.Game
		sc   = game.ShowCount.Int()
	)

	if len(game.Odais) < sc+1 {
		return errNotFound
	}

	var next oapi.WsNextShowStatus
	if sc+1 < len(game.Odais) {
		next = oapi.WsNextShowStatus("odai")
	} else {
		next = oapi.WsNextShowStatus("end")
	}

	answerer := oapi.User{}
	for _, m := range c.room.Members {
		if m.Id == game.Odais[sc].AnswererId {
			answerer = oapi.RefillUser(&m)
			break
		}
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWANSWER,
			Body: &oapi.WsShowAnswerEventBody{
				Answerer: answerer,
				Next:     next,
				Answer:   game.Odais[sc].Answer.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	if c.userId == c.room.HostId {
		if sc+1 < len(game.Odais) {
			c.room.Game.NextShowPhase = model.GameShowPhaseOdai
		} else {
			c.room.Game.NextShowPhase = model.GameShowPhaseEnd
		}

		c.room.Game.ShowCount++
	}

	c.sendMsg(buf)

	return nil
}

// RETURN_ROOM
// ルーム(新規加入待機状態) に戻る (ホスト -> サーバー)
// このタイミングでサーバーは保持しているゲームデータを削除
func (c *Client) receiveReturnRoomEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	c.room.ResetGame()
	c.resetBreakTimer()

	c.bloadcast(func(cc *Client) {
		if err := cc.sendNextRoomEvent(); err != nil {
			log.Println("failed to send NEXT_ROOM event:", err.Error())
		}
	})

	return nil
}

// NEXT_ROOM
// ルームの表示に遷移する (サーバー -> ルーム全員)
func (c *Client) sendNextRoomEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventNEXTROOM,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.sendMsg(buf)

	return nil
}

// CHANGE_HOST
// ホストが落ちた時に飛んできて，ホスト役を変更する (サーバー -> ルーム全員)
func (c *Client) sendChangeHostEvent() error {
	room := c.room
	found := false
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	for _, v := range room.Members {
		if _, ok := c.hub.userIdToClient[v.Id]; ok && v.Id != room.HostId {
			found = true
			room.HostId = v.Id
			break
		}
	}

	if !found {
		return errNotFound
	}

	return nil
}

// BREAK_ROOM
// 部屋が破壊されたときに通知する (サーバー -> ルーム全員)
// 部屋が立ってからゲーム開始まで15分以上経過している場合，部屋を閉じる
// このタイミングでサーバーは保持しているルームに関わる全データを削除
func (c *Client) sendBreakRoomEvent() error {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	for _, v := range c.room.Members {
		if v.Id == c.userId {
			continue
		}

		if cc, ok := c.hub.userIdToClient[v.Id]; ok {
			c.hub.unregister(cc)
		}
	}

	if err := c.hub.repo.DeleteRoom(c.room.Id); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	c.hub.unregister(c)

	return nil
}
