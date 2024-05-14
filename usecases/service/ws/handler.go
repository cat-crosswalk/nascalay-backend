package ws

import (
	"fmt"
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
	oapi.WsEventROOMSETOPTION:    (*Client).sendRoomSetOptionEvent,
	oapi.WsEventREQUESTGAMESTART: (*Client).sendRequestGameStartEvent,
	oapi.WsEventODAIREADY:        (*Client).sendOdaiReadyEvent,
	oapi.WsEventODAICANCEL:       (*Client).sendOdaiCancelEvent,
	oapi.WsEventODAISEND:         (*Client).sendOdaiSendEvent,
	oapi.WsEventDRAWREADY:        (*Client).sendDrawReadyEvent,
	oapi.WsEventDRAWCANCEL:       (*Client).sendDrawCancelEvent,
	oapi.WsEventDRAWSEND:         (*Client).sendDrawSendEvent,
	oapi.WsEventANSWERREADY:      (*Client).sendAnswerReadyEvent,
	oapi.WsEventANSWERCANCEL:     (*Client).sendAnswerCancelEvent,
	oapi.WsEventANSWERSEND:       (*Client).sendAnswerSendEvent,
	oapi.WsEventSHOWNEXT:         (*Client).sendShowNextEvent,
	oapi.WsEventRETURNROOM:       (*Client).sendReturnRoomEvent,
}

// ROOM_NEW_MEMBER
// 部屋に追加のメンバーが来たことを通知する (サーバー -> ルーム全員)
func (s *RoomServer) sendRoomNewMemberEvent(room *model.Room) error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
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
func (c *Client) sendRoomSetOptionEvent(body interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	if c.userId != c.server.room.HostId {
		return errUnAuthorized
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsRoomSetOptionEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	updateBody := new(oapi.WsRoomUpdateOptionEventBody)
	game := c.server.room.Game

	// Set options
	if e.TimeLimit != nil {
		game.TimeLimit = model.TimeLimit(*e.TimeLimit)
		updateBody.TimeLimit = e.TimeLimit
	}

	if err := c.server.sendRoomUpdateOptionEvent(updateBody); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventROOMUPDATEOPTION)
	}

	return nil
}

// ROOM_UPDATE_OPTION
// ゲームの設定を更新する (サーバー -> ルーム全員)
func (s *RoomServer) sendRoomUpdateOptionEvent(body *oapi.WsRoomUpdateOptionEventBody) error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventROOMUPDATEOPTION,
		Body: body,
	})

	return nil
}

// REQUEST_GAME_START
// ゲームを開始する (ホスト -> サーバー)
func (c *Client) sendRequestGameStartEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}
	if len(c.server.room.Members) < 2 {
		return errNotEnoughMember
	}

	if stopped := c.server.room.Game.BreakTimer.Stop(); !stopped {
		go c.server.waitAndBreakRoom()
	}

	if c.userId != c.server.room.HostId {
		return errUnAuthorized
	}

	if err := c.server.sendGameStartEvent(); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventGAMESTART)
	}

	return nil
}

// GAME_START
// ゲームの開始を通知する (サーバー -> ルーム全員)
// ODAIフェーズを開始する
func (s *RoomServer) sendGameStartEvent() error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	for _, m := range s.room.Members {
		c, ok := s.hub.userIdToClient[m.Id]
		if !ok {
			c.logger.Infof("client(userId:%s) not found", m.Id.UUID().String())
			continue
		}

		s.sendMsgTo(c, &oapi.WsSendMessage{
			Type: oapi.WsEventGAMESTART,
			Body: &oapi.WsGameStartEventBody{
				OdaiExample: random.OdaiExample(),
				TimeLimit:   int(s.room.Game.TimeLimit),
			},
		})
	}

	// ODAIフェーズに移行
	s.room.Game.Status = model.GameStatusOdai
	// ODAIのカウントダウン開始
	if !s.room.Game.Timer.Stop() {
		<-s.room.Game.Timer.C
	}
	s.room.Game.Timer.Reset(time.Second * time.Duration(s.room.Game.TimeLimit))
	s.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(s.room.Game.TimeLimit)))

	go func() {
		select {
		case <-s.room.Game.Timer.C:
			if err := s.sendOdaiFinishEvent(); err != nil {
				s.logger.Error(s.sendEventErr(err, oapi.WsEventODAIFINISH).Error())
			}
		case <-s.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ODAI_READY
// お題の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) sendOdaiReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.server.allMembersAreReady() {
		if c.server.room.Game.Timer.Stop() {
			c.server.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.server.room.Game.Timer.C
		}
		if err := c.server.sendOdaiFinishEvent(); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventODAIFINISH)
		}
	} else {
		if err := c.server.sendOdaiInputEvent(len(c.server.room.Game.Ready)); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventODAIINPUT)
		}
	}

	return nil
}

// ODAI_CANCEL
// お題の入力の完了を解除する (ルームの各員 -> サーバー)
func (c *Client) sendOdaiCancelEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	if err := c.server.sendOdaiInputEvent(len(c.server.room.Game.Ready)); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventODAIINPUT)
	}

	return nil
}

// ODAI_INPUT
// お題入力が完了した人数を送信する (サーバー -> ルームの各員)
func (s *RoomServer) sendOdaiInputEvent(readyNum int) error {
	if !s.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventODAIINPUT,
		Body: &oapi.WsOdaiInputEventBody{
			Ready: readyNum,
		},
	})

	return nil
}

// ODAI_FINISH
// 全員がお題の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントはお題を送信する
func (s *RoomServer) sendOdaiFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventODAIFINISH,
	})

	return nil
}

// ODAI_SEND
// お題を送信する (ルームの各員 -> サーバー)
// DRAWフェーズを開始する
func (c *Client) sendOdaiSendEvent(body interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsOdaiSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	game := c.server.room.Game

	// 存在チェック
	for _, v := range game.Odais {
		if v.SenderId == c.userId || v.Title == model.OdaiTitle(e.Odai) {
			return errAlreadyExists
		}
	}

	game.AddOdai(c.userId, model.OdaiTitle(e.Odai))

	// 全員のお題送信が完了したらDRAWフェーズに移行
	odaisByUnregisteredClients := make([]model.UserId, 0, len(c.server.room.Members)) // ハブから登録解除したクライアントの配列
	for _, v := range c.server.room.Members {
		if _, ok := c.hub.userIdToClient[v.Id]; !ok {
			odaisByUnregisteredClients = append(odaisByUnregisteredClients, v.Id)
		}
	}

	if len(game.Odais)+len(odaisByUnregisteredClients) == len(c.server.room.Members) {
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
		random.SetupMemberRoles(game, c.server.room.Members)

		if err := c.server.sendDrawStartEvent(); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventDRAWSTART)
		}

		// DRAWのカウントダウン開始
		game.Timer.Reset(time.Second * time.Duration(c.server.room.Game.TimeLimit))
		game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.server.room.Game.TimeLimit)))

		go func() {
			select {
			case <-game.Timer.C:
				if err := c.server.sendDrawFinishEvent(); err != nil {
					c.logger.Error(c.server.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
				}
			case <-game.TimerStopChan:
			}
		}()
	}

	return nil
}

// DRAW_START
// キャンバス情報とお題を送信する (サーバー -> ルーム各員)
func (s *RoomServer) sendDrawStartEvent() error {
	// NOTE: 最後にお題を送信したユーザー(2回目以降は最後に絵を送信したユーザー)のクライアントから実行される
	if !s.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	var (
		game      = s.room.Game
		drawCount = game.DrawCount
	)

	for _, o := range game.Odais {
		if len(o.DrawerSeq) <= drawCount.Int() {
			return errInvalidDrawCount
		}

		drawer := o.DrawerSeq[drawCount.Int()]
		c, ok := s.hub.userIdToClient[drawer.UserId] // `drawCount`番目に描くユーザーのクライアント
		if !ok {
			s.logger.Infof("client(userId:%s) not found", drawer.UserId.UUID().String())
			continue
		}

		drawnArea := make([]int, drawCount.Int())
		for i := 0; i < drawCount.Int(); i++ {
			drawnArea[i] = o.DrawerSeq[i].AreaId.Int()
		}

		s.sendMsgTo(c, &oapi.WsSendMessage{
			Type: oapi.WsEventDRAWSTART,
			Body: oapi.WsDrawStartEventBody{
				AllDrawPhaseNum: s.room.AllDrawPhase(),
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
func (c *Client) sendDrawReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.server.allMembersAreReady() {
		if c.server.room.Game.Timer.Stop() {
			c.server.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.server.room.Game.Timer.C
		}
		if err := c.server.sendDrawFinishEvent(); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventDRAWFINISH)
		}
	} else {
		if err := c.server.sendDrawInputEvent(len(c.server.room.Game.Ready)); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventDRAWINPUT)
		}
	}

	return nil
}

// DRAW_CANCEL
// 絵が書き終わっている通知を解除する (ルームの各員 -> サーバー)
func (c *Client) sendDrawCancelEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	if err := c.server.sendDrawInputEvent(len(c.server.room.Game.Ready)); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventDRAWINPUT)
	}

	return nil
}

// DRAW_INPUT
// 絵を描き終えた人数を送信する (サーバー -> ルームの各員)
func (s *RoomServer) sendDrawInputEvent(readyNum int) error {
	if !s.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventDRAWINPUT,
		Body: &oapi.WsDrawInputEventBody{
			Ready: readyNum,
		},
	})

	return nil
}

// DRAW_FINISH
// 全員が絵を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは絵を送信する
func (s *RoomServer) sendDrawFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventDRAWFINISH,
	})

	return nil
}

// DRAW_SEND
// 絵を送信する (ルームの各員 -> サーバー)
// お題が残っていたら再度DRAW_START が送信される
// お題がすべて終わったらANSWERフェーズを開始する
func (c *Client) sendDrawSendEvent(body interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsDrawSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	for _, v := range c.server.room.Game.Odais {
		if v.DrawerSeq[c.server.room.Game.DrawCount].UserId == c.userId {
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
	for _, v := range c.server.room.Game.Odais {
		if _, ok := c.hub.userIdToClient[v.DrawerSeq[c.server.room.Game.DrawCount.Int()].UserId]; !ok {
			continue
		}

		if !v.ImgUpdated {
			allImgUpdated = false
			break
		}
	}

	game := c.server.room.Game
	if allImgUpdated {
		game.ResetReady()
		if game.DrawCount.Int()+1 < c.server.room.AllDrawPhase() {
			game.DrawCount++

			game.ResetImgUpdated()

			if err := c.server.sendDrawStartEvent(); err != nil {
				return c.server.sendEventErr(err, oapi.WsEventDRAWSTART)
			}

			// DRAWのカウントダウン開始
			c.server.room.Game.Timer.Reset(time.Second * time.Duration(c.server.room.Game.TimeLimit))
			c.server.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(c.server.room.Game.TimeLimit)))

			go func() {
				select {
				case <-c.server.room.Game.Timer.C:
					if err := c.server.sendDrawFinishEvent(); err != nil {
						c.logger.Error(c.server.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
					}
				case <-c.server.room.Game.TimerStopChan:
				}
			}()
		} else {
			game.Status = model.GameStatusAnswer

			if err := c.server.sendAnswerStartEvent(); err != nil {
				return c.server.sendEventErr(err, oapi.WsEventANSWERSTART)
			}
		}
	}

	return nil
}

// ANSWER_START
// 絵が飛んできて，回答する (サーバー -> ルーム各員)
func (s *RoomServer) sendAnswerStartEvent() error {
	// NOTE: 最後にお題を送信した人のクライアントで行う
	if !s.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	for _, v := range s.room.Game.Odais {
		ac, ok := s.hub.userIdToClient[v.AnswererId]
		if !ok {
			s.logger.Infof("client(userId:%s) not found", v.AnswererId.UUID().String())
			continue
		}

		s.sendMsgTo(ac, &oapi.WsSendMessage{
			Type: oapi.WsEventANSWERSTART,
			Body: oapi.WsAnswerStartEventBody{
				Img:       v.Img.AddPrefix(),
				TimeLimit: int(s.room.Game.TimeLimit),
			},
		})
	}

	// ANSWERのカウントダウン開始
	s.room.Game.Timer.Reset(time.Second * time.Duration(s.room.Game.TimeLimit))
	s.room.Game.Timeout = model.Timeout(time.Now().Add(time.Second * time.Duration(s.room.Game.TimeLimit)))

	go func() {
		select {
		case <-s.room.Game.Timer.C:
			if err := s.sendAnswerFinishEvent(); err != nil {
				s.logger.Error(s.sendEventErr(err, oapi.WsEventANSWERFINISH).Error())
			}
		case <-s.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ANSWER_READY
// 回答の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) sendAnswerReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.AddReady(c.userId)

	if c.server.allMembersAreReady() {
		if c.server.room.Game.Timer.Stop() {
			c.server.room.Game.TimerStopChan <- struct{}{}
		} else {
			<-c.server.room.Game.Timer.C
		}
		if err := c.server.sendAnswerFinishEvent(); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventANSWERFINISH)
		}
	} else {
		if err := c.server.sendAnswerInputEvent(len(c.server.room.Game.Ready)); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventANSWERINPUT)
		}
	}

	return nil
}

// ANSWER_CANCEL
// 回答の入力の完了を解除する (ルームの各員 -> サーバー)
func (c *Client) sendAnswerCancelEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.CancelReady(c.userId)

	if err := c.server.sendAnswerInputEvent(len(c.server.room.Game.Ready)); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventANSWERINPUT)
	}

	return nil
}

// ANSWER_INPUT
// 回答の入力が完了した人数を送信する (サーバー -> ルームの各員)
func (s *RoomServer) sendAnswerInputEvent(readyNum int) error {
	if !s.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventANSWERINPUT,
		Body: &oapi.WsAnswerInputEventBody{
			Ready: readyNum,
		},
	})

	return nil
}

// ANSWER_FINISH
// 全員が回答の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは回答を送信する
func (s *RoomServer) sendAnswerFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventANSWERFINISH,
	})

	return nil
}

// ANSWER_SEND
// 回答を送信する (ルームの各員 -> サーバー)
// SHOWフェーズを開始する
func (c *Client) sendAnswerSendEvent(body interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	if body == nil {
		return errNilBody
	}

	e := new(oapi.WsAnswerSendEventBody)
	if err := mapstructure.Decode(body, e); err != nil {
		return fmt.Errorf("failed to decode body: %w", err)
	}

	game := c.server.room.Game

	for _, v := range game.Odais {
		if v.AnswererId == c.userId {
			ma := model.OdaiAnswer(e.Answer)
			v.Answer = &ma
			break
		}
	}

	// 全員の回答が送信されたらSHOWフェーズに移行
	allAnswersendd := true
	for _, v := range game.Odais {
		if _, ok := c.hub.userIdToClient[v.AnswererId]; !ok {
			continue
		}

		if v.Answer == nil {
			allAnswersendd = false
			break
		}
	}

	if allAnswersendd {
		game.Status = model.GameStatusShow

		if err := c.server.sendShowStartEvent(); err != nil {
			return c.server.sendEventErr(err, oapi.WsEventSHOWSTART)
		}
	}

	return nil
}

// SHOW_START
// 結果表示フェーズが始まったことを通知する (サーバー -> ルーム全員)
func (s *RoomServer) sendShowStartEvent() error {
	// NOTE: 最後に回答したユーザーのクライアントで行う
	if !s.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	s.resetBreakTimer()

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventSHOWSTART,
	})

	return nil
}

// SHOW_NEXT
// つぎの結果表示を要求する (ホスト -> サーバー)
func (c *Client) sendShowNextEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	if c.userId != c.server.room.HostId {
		return errUnAuthorized
	}

	game := c.server.room.Game

	switch game.NextShowPhase {
	case model.GameShowPhaseOdai:
		if err := c.server.sendShowOdaiEvent(); err != nil {
			c.logger.Error(c.server.sendEventErr(err, oapi.WsEventSHOWODAI))
		}
		game.NextShowPhase = model.GameShowPhaseCanvas
	case model.GameShowPhaseCanvas:
		if err := c.server.sendShowCanvasEvent(); err != nil {
			c.logger.Error(c.server.sendEventErr(err, oapi.WsEventSHOWCANVAS))
		}
		game.NextShowPhase = model.GameShowPhaseAnswer
	case model.GameShowPhaseAnswer:
		if err := c.server.sendShowAnswerEvent(); err != nil {
			c.logger.Error(c.server.sendEventErr(err, oapi.WsEventSHOWANSWER))
		}
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
func (s *RoomServer) sendShowOdaiEvent() error {
	if !s.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}
	var (
		game = s.room.Game
		sc   = game.ShowCount.Int()
	)
	if len(game.Odais) < sc+1 {
		return errNotFound
	}

	sender := oapi.User{}
	for _, m := range s.room.Members {
		if m.Id == game.Odais[sc].SenderId {
			sender = oapi.RefillUser(&m)
			break
		}
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
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
func (s *RoomServer) sendShowCanvasEvent() error {
	if !s.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	var (
		game = s.room.Game
		sc   = game.ShowCount.Int()
	)

	if len(game.Odais) < sc+1 {
		return errNotFound
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
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
func (s *RoomServer) sendShowAnswerEvent() error {
	if !s.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	var (
		game = s.room.Game
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
	for _, m := range s.room.Members {
		if m.Id == game.Odais[sc].AnswererId {
			answerer = oapi.RefillUser(&m)
			break
		}
	}

	var answer string
	if a := game.Odais[sc].Answer; a != nil {
		answer = a.String()
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
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
func (c *Client) sendReturnRoomEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	if c.userId != c.server.room.HostId {
		return errUnAuthorized
	}

	c.server.room.ResetGame()
	c.server.resetBreakTimer()

	if err := c.server.sendNextRoomEvent(); err != nil {
		c.logger.Error(c.server.sendEventErr(err, oapi.WsEventNEXTROOM))
	}

	return nil
}

// NEXT_ROOM
// ルームの表示に遷移する (サーバー -> ルーム全員)
func (s *RoomServer) sendNextRoomEvent() error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventNEXTROOM,
	})

	return nil
}

// CHANGE_HOST
// ホストが落ちた時に飛んできて，ホスト役を変更する (サーバー -> ルーム全員)
func (s *RoomServer) sendChangeHostEvent() error {
	room := s.room
	found := false
	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	for _, v := range room.Members {
		if _, ok := s.hub.userIdToClient[v.Id]; ok && v.Id != room.HostId {
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
func (s *RoomServer) sendBreakRoomEvent() error {
	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	for _, v := range s.room.Members {
		if c, ok := s.hub.userIdToClient[v.Id]; ok {
			s.hub.unregister(c)
		}
	}

	if err := s.hub.repo.DeleteRoom(s.room.Id); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	return nil
}
