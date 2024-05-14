package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/util/logger"
	"github.com/21hack02win/nascalay-backend/util/random"
)

type Server struct {
	hub  *Hub
	room *model.Room
}

// ROOM_NEW_MEMBER
// 部屋に追加のメンバーが来たことを通知する (サーバー -> ルーム全員)
func (s *Server) sendRoomNewMemberEvent(room *model.Room) error {
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

// ROOM_UPDATE_OPTION
// ゲームの設定を更新する (サーバー -> ルーム全員)
func (s *Server) sendRoomUpdateOptionEvent(body *oapi.WsRoomUpdateOptionEventBody) error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventROOMUPDATEOPTION,
		Body: body,
	})

	return nil
}

// GAME_START
// ゲームの開始を通知する (サーバー -> ルーム全員)
// ODAIフェーズを開始する
func (s *Server) sendGameStartEvent() error {
	if !s.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	for _, m := range s.room.Members {
		c, ok := s.hub.userIdToClient[m.Id]
		if !ok {
			logger.Echo.Infof("client(userId:%s) not found", m.Id.UUID().String())
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
				logger.Echo.Error(s.sendEventErr(err, oapi.WsEventODAIFINISH).Error())
			}
		case <-s.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ODAI_INPUT
// お題入力が完了した人数を送信する (サーバー -> ルームの各員)
func (s *Server) sendOdaiInputEvent(readyNum int) error {
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
func (s *Server) sendOdaiFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventODAIFINISH,
	})

	return nil
}

// DRAW_START
// キャンバス情報とお題を送信する (サーバー -> ルーム各員)
func (s *Server) sendDrawStartEvent() error {
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
			logger.Echo.Infof("client(userId:%s) not found", drawer.UserId.UUID().String())
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

// DRAW_INPUT
// 絵を描き終えた人数を送信する (サーバー -> ルームの各員)
func (s *Server) sendDrawInputEvent(readyNum int) error {
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
func (s *Server) sendDrawFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventDRAWFINISH,
	})

	return nil
}

// ANSWER_START
// 絵が飛んできて，回答する (サーバー -> ルーム各員)
func (s *Server) sendAnswerStartEvent() error {
	// NOTE: 最後にお題を送信した人のクライアントで行う
	if !s.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	for _, v := range s.room.Game.Odais {
		ac, ok := s.hub.userIdToClient[v.AnswererId]
		if !ok {
			logger.Echo.Infof("client(userId:%s) not found", v.AnswererId.UUID().String())
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
				logger.Echo.Error(s.sendEventErr(err, oapi.WsEventANSWERFINISH).Error())
			}
		case <-s.room.Game.TimerStopChan:
		}
	}()

	return nil
}

// ANSWER_INPUT
// 回答の入力が完了した人数を送信する (サーバー -> ルームの各員)
func (s *Server) sendAnswerInputEvent(readyNum int) error {
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
func (s *Server) sendAnswerFinishEvent() error {
	if !s.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	s.sendMsgToEachClientInRoom(&oapi.WsSendMessage{
		Type: oapi.WsEventANSWERFINISH,
	})

	return nil
}

// SHOW_START
// 結果表示フェーズが始まったことを通知する (サーバー -> ルーム全員)
func (s *Server) sendShowStartEvent() error {
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

// SHOW_ODAI
// 最初のお題を受信する (サーバー -> ルーム全員)
func (s *Server) sendShowOdaiEvent() error {
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
func (s *Server) sendShowCanvasEvent() error {
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
func (s *Server) sendShowAnswerEvent() error {
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

// NEXT_ROOM
// ルームの表示に遷移する (サーバー -> ルーム全員)
func (s *Server) sendNextRoomEvent() error {
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
func (s *Server) sendChangeHostEvent() error {
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
func (s *Server) sendBreakRoomEvent() error {
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

// Utils

// Send message to a client
func (s *Server) sendMsgTo(c *Client, msg *oapi.WsSendMessage) {
	// If the client is not connected, remove the client from the room
	// If the client is the host, change the host
	if c.send == nil {
		if c.userId == s.room.HostId {
			if err := s.sendChangeHostEvent(); err != nil {
				logger.Echo.Error(s.sendEventErr(err, oapi.WsEventCHANGEHOST))
			}
		}

		s.hub.unregister(c)

		return
	}

	c.send <- msg
}

// Send message to all clients in the room
func (s *Server) sendMsgToEachClientInRoom(msg *oapi.WsSendMessage) {
	var wg sync.WaitGroup
	for _, m := range s.room.Members {
		c, ok := s.hub.userIdToClient[m.Id]
		if !ok {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.sendMsgTo(c, msg)
		}()
	}
	wg.Wait()
}

// Check if all members are ready
func (s *Server) allMembersAreReady() bool {
	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	r := s.room
	for _, m := range r.Members {
		if _, ok := s.hub.userIdToClient[m.Id]; !ok {
			continue
		}

		if _, ok := r.Game.Ready[m.Id]; !ok {
			return false
		}
	}

	return true
}

// Reset break timer
func (s *Server) resetBreakTimer() {
	// タイマーをリセットする
	// 15(分)後に次のゲームが始まらなければルームを削除する
	game := s.room.Game
	bt := game.BreakTimer

	if stopped := bt.Stop(); !stopped {
		go s.waitAndBreakRoom()
	}

	bt.Reset(time.Minute * 15)
	go s.waitAndBreakRoom()
}

// Wait for 15 minutes and break the room
func (s *Server) waitAndBreakRoom() {
	<-s.room.Game.BreakTimer.C
	if err := s.sendBreakRoomEvent(); err != nil {
		logger.Echo.Error(s.sendEventErr(err, oapi.WsEventBREAKROOM))
	}
}

func (s *Server) sendEventErr(err error, eventName oapi.WsEvent) error {
	return fmt.Errorf(
		"[ERROR] failed to send %s event (roomId:%s): %w",
		eventName,
		s.room.Id.String(),
		err,
	)
}
