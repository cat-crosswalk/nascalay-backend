package ws

import (
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

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventROOMNEWMEMBER,
		Body: oapi.WsRoomNewMemberEventBody{
			Capacity: room.Capacity.Int(),
			HostId:   room.HostId.UUID(),
			Members:  oapi.RefillUsers(room.Members),
		},
	})

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
		return c.sendEventErr(err, oapi.WsEventROOMUPDATEOPTION)
	}

	return nil
}

// ROOM_UPDATE_OPTION
// ゲームの設定を更新する (サーバー -> ルーム全員)
func (c *Client) sendRoomUpdateOptionEvent(body *oapi.WsRoomUpdateOptionEventBody) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventROOMUPDATEOPTION,
		Body: body,
	})

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
		return c.sendEventErr(err, oapi.WsEventGAMESTART)
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
		cc, ok := c.hub.userIdToClient[m.Id]
		if !ok {
			log.Printf("[INFO] client(userId:%s) not found", m.Id.UUID().String())
			continue
		}

		cc.sendMsg(&oapi.WsSendMessage{
			Type: oapi.WsEventGAMESTART,
			Body: &oapi.WsGameStartEventBody{
				OdaiExample: random.OdaiExample(),
				TimeLimit:   int(c.room.Game.TimeLimit),
			},
		})
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
				log.Println(c.sendEventErr(err, oapi.WsEventODAIFINISH).Error())
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
			return c.sendEventErr(err, oapi.WsEventODAIFINISH)
		}
	} else {
		if err := c.sendOdaiInputEvent(); err != nil {
			return c.sendEventErr(err, oapi.WsEventODAIINPUT)
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

	if err := c.sendOdaiInputEvent(); err != nil {
		return c.sendEventErr(err, oapi.WsEventODAIINPUT)
	}

	return nil
}

// ODAI_INPUT
// お題入力が完了した人数を送信する (サーバー -> ルームの各員)
func (c *Client) sendOdaiInputEvent() error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventODAIINPUT,
		Body: &oapi.WsOdaiInputEventBody{
			Ready: len(c.room.Game.Ready),
		},
	})

	return nil
}

// ODAI_FINISH
// 全員がお題の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントはお題を送信する
func (c *Client) sendOdaiFinishEvent() error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventODAIFINISH,
	})

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

	game := c.room.Game

	// 存在チェック
	for _, v := range game.Odais {
		if v.SenderId == c.userId || v.Title == model.OdaiTitle(e.Odai) {
			return errAlreadyExists
		}
	}

	game.AddOdai(c.userId, model.OdaiTitle(e.Odai))

	// 全員のお題送信が完了したらDRAWフェーズに移行
	odaisByUnregisteredClients := make([]model.UserId, 0, len(c.room.Members)) // ハブから登録解除したクライアントの配列
	for _, v := range c.room.Members {
		if _, ok := c.hub.userIdToClient[v.Id]; !ok {
			odaisByUnregisteredClients = append(odaisByUnregisteredClients, v.Id)
		}
	}

	if len(game.Odais)+len(odaisByUnregisteredClients) == len(c.room.Members) {
		for _, Id := range odaisByUnregisteredClients {
			game.Odais = append(game.Odais, &model.Odai{
				Title:     model.OdaiTitle(random.OdaiExample()),
				SenderId:  Id,
				DrawerSeq: []model.Drawer{},
			})
		}
		game.ResetReady()
		game.Status = model.GameStatusDraw
		game.DrawCount = 0
		random.SetupMemberRoles(game, c.room.Members)

		if err := c.sendDrawStartEvent(); err != nil {
			return c.sendEventErr(err, oapi.WsEventDRAWSTART)
		}

		// DRAWのカウントダウン開始
		game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
		game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

		go func() {
			select {
			case <-game.Timer.C:
				if err := c.sendDrawFinishEvent(); err != nil {
					log.Println(c.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
				}
			case <-game.TimerStopChan:
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
			log.Printf("[INFO] client(userId:%s) not found", drawer.UserId.UUID().String())
			continue
		}

		drawnArea := make([]int, drawCount.Int())
		for i := 0; i < drawCount.Int(); i++ {
			drawnArea[i] = o.DrawerSeq[i].AreaId.Int()
		}

		cc.sendMsg(&oapi.WsSendMessage{
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
		})
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
			return c.sendEventErr(err, oapi.WsEventDRAWFINISH)
		}
	} else {
		if err := c.sendDrawInputEvent(); err != nil {
			return c.sendEventErr(err, oapi.WsEventDRAWINPUT)
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

	if err := c.sendDrawInputEvent(); err != nil {
		return c.sendEventErr(err, oapi.WsEventDRAWINPUT)
	}

	return nil
}

// DRAW_INPUT
// 絵を描き終えた人数を送信する (サーバー -> ルームの各員)
func (c *Client) sendDrawInputEvent() error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventDRAWINPUT,
		Body: &oapi.WsDrawInputEventBody{
			Ready: len(c.room.Game.Ready),
		},
	})

	return nil
}

// DRAW_FINISH
// 全員が絵を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは絵を送信する
func (c *Client) sendDrawFinishEvent() error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventDRAWFINISH,
	})

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
		if _, ok := c.hub.userIdToClient[v.DrawerSeq[c.room.Game.DrawCount.Int()].UserId]; !ok {
			continue
		}

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
				return c.sendEventErr(err, oapi.WsEventDRAWSTART)
			}

			// DRAWのカウントダウン開始
			c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
			c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

			go func() {
				select {
				case <-c.room.Game.Timer.C:
					if err := c.sendDrawFinishEvent(); err != nil {
						log.Println(c.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
					}
				case <-c.room.Game.TimerStopChan:
				}
			}()
		} else {
			game.Status = model.GameStatusAnswer

			if err := c.sendAnswerStartEvent(); err != nil {
				return c.sendEventErr(err, oapi.WsEventANSWERSTART)
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
			log.Printf("[INFO] client(userId:%s) not found", v.AnswererId.UUID().String())
			continue
		}

		ac.sendMsg(&oapi.WsSendMessage{
			Type: oapi.WsEventANSWERSTART,
			Body: oapi.WsAnswerStartEventBody{
				Img:       v.Img.AddPrefix(),
				TimeLimit: int(c.room.Game.TimeLimit),
			},
		})
	}

	// ANSWERのカウントダウン開始
	c.room.Game.Timer.Reset(time.Second * time.Duration(c.room.Game.TimeLimit))
	c.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.room.Game.TimeLimit)))

	go func() {
		select {
		case <-c.room.Game.Timer.C:
			if err := c.sendAnswerFinishEvent(); err != nil {
				log.Println(c.sendEventErr(err, oapi.WsEventANSWERFINISH).Error())
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
			return c.sendEventErr(err, oapi.WsEventANSWERFINISH)
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

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventANSWERFINISH,
	})

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
		if _, ok := c.hub.userIdToClient[v.AnswererId]; !ok {
			continue
		}

		if v.Answer == nil {
			allAnswerReceived = false
			break
		}
	}

	if allAnswerReceived {
		game.Status = model.GameStatusShow

		if err := c.sendShowStartEvent(); err != nil {
			return c.sendEventErr(err, oapi.WsEventSHOWSTART)
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

	c.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventSHOWSTART,
	})

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

	game := c.room.Game

	switch game.NextShowPhase {
	case model.GameShowPhaseOdai:
		c.broadcast(func(cc *Client) {
			if err := cc.sendShowOdaiEvent(); err != nil {
				log.Println(cc.sendEventErr(err, oapi.WsEventSHOWODAI))
			}
		})
		game.NextShowPhase = model.GameShowPhaseCanvas
	case model.GameShowPhaseCanvas:
		c.broadcast(func(cc *Client) {
			if err := cc.sendShowCanvasEvent(); err != nil {
				log.Println(cc.sendEventErr(err, oapi.WsEventSHOWCANVAS))
			}
		})
		game.NextShowPhase = model.GameShowPhaseAnswer
	case model.GameShowPhaseAnswer:
		c.broadcast(func(cc *Client) {
			if err := cc.sendShowAnswerEvent(); err != nil {
				log.Println(cc.sendEventErr(err, oapi.WsEventSHOWANSWER))
			}
		})
		if game.ShowCount.Int()+1 < len(game.Odais) {
			game.NextShowPhase = model.GameShowPhaseOdai
		} else {
			game.NextShowPhase = model.GameShowPhaseEnd
		}
		game.ShowCount++
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

	c.sendMsg(&oapi.WsSendMessage{
		Type: oapi.WsEventSHOWODAI,
		Body: &oapi.WsShowOdaiEventBody{
			Sender: sender,
			Next:   oapi.WsNextShowStatus("canvas"),
			Odai:   game.Odais[sc].Title.String(),
		},
	})

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

	c.sendMsg(&oapi.WsSendMessage{
		Type: oapi.WsEventSHOWCANVAS,
		Body: &oapi.WsShowCanvasEventBody{
			Next: oapi.WsNextShowStatus("answer"),
			Img:  game.Odais[sc].Img.AddPrefix(),
		},
	})

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

	var answer string
	if a := game.Odais[sc].Answer; a != nil {
		answer = a.String()
	}

	c.sendMsg(&oapi.WsSendMessage{
		Type: oapi.WsEventSHOWANSWER,
		Body: &oapi.WsShowAnswerEventBody{
			Answerer: answerer,
			Next:     next,
			Answer:   answer,
		},
	})

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

	c.broadcast(func(cc *Client) {
		if err := cc.sendNextRoomEvent(); err != nil {
			log.Println(cc.sendEventErr(err, oapi.WsEventNEXTROOM))
		}
	})

	return nil
}

// NEXT_ROOM
// ルームの表示に遷移する (サーバー -> ルーム全員)
func (c *Client) sendNextRoomEvent() error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	c.sendMsg(&oapi.WsSendMessage{
		Type: oapi.WsEventNEXTROOM,
	})

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
