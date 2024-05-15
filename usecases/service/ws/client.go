//nolint:errcheck
package ws

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/util/canvas"
	"github.com/21hack02win/nascalay-backend/util/logger"
	"github.com/21hack02win/nascalay-backend/util/random"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 300000
)

type Client struct {
	hub    *Hub
	userId model.UserId
	server *Server
	conn   *websocket.Conn
	send   chan *oapi.WsSendMessage
}

func NewClient(hub *Hub, userId model.UserId, conn *websocket.Conn) (*Client, error) {
	room, err := hub.repo.GetRoomFromUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get room from userId: %w", err)
	}

	server, ok := hub.roomIdToServer[room.Id]
	if !ok {
		server = &Server{
			hub:  hub,
			room: room,
		}
		hub.mux.Lock()
		hub.roomIdToServer[room.Id] = server
		hub.mux.Unlock()
	}

	return &Client{
		hub:    hub,
		userId: userId,
		server: server,
		conn:   conn,
		send:   make(chan *oapi.WsSendMessage, 256),
	}, nil
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Echo.Error("failed to create next writer:", err.Error())
				return
			}

			buf, err := json.Marshal(message)
			if err != nil {
				logger.Echo.Error("failed to encode as JSON:", err.Error())
				return
			}

			w.Write(buf)

			// Add queued chat messages to the current websocket message.
			for i := 0; i < len(c.send); i++ {
				buf, err = json.Marshal(<-c.send)
				if err != nil {
					logger.Echo.Error("failed to encode as JSON:", err.Error())
					return
				}

				w.Write(buf)
			}

			if err := w.Close(); err != nil {
				logger.Echo.Error("failed to close writer:", err.Error())
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Echo.Error("failed to write message:", err.Error())
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregisterCh <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		req := new(oapi.WsJSONRequestBody)
		if err := c.conn.ReadJSON(req); err != nil {
			if !websocket.IsCloseError(err) && !websocket.IsUnexpectedCloseError(err) {
				logger.Echo.Error("websocket error occured:", err.Error())
			}
			break
		}

		if err := c.callEventHandler(req); err != nil {
			logger.Echo.Error("websocket error occured:", err.Error())
			c.send <- &oapi.WsSendMessage{
				Type: oapi.WsEventERROR,
				Body: &oapi.WsErrorBody{
					Content: err.Error(),
				},
			}
			continue
		}
	}
}

// Client Events

func (c *Client) callEventHandler(req *oapi.WsJSONRequestBody) error {
	switch req.Type {
	case oapi.WsEventROOMSETOPTION:
		return c.sendRoomSetOptionEvent(req.Body)
	case oapi.WsEventREQUESTGAMESTART:
		return c.sendRequestGameStartEvent(req.Body)
	case oapi.WsEventODAIREADY:
		return c.sendOdaiReadyEvent(req.Body)
	case oapi.WsEventODAICANCEL:
		return c.sendOdaiCancelEvent(req.Body)
	case oapi.WsEventODAISEND:
		return c.sendOdaiSendEvent(req.Body)
	case oapi.WsEventDRAWREADY:
		return c.sendDrawReadyEvent(req.Body)
	case oapi.WsEventDRAWCANCEL:
		return c.sendDrawCancelEvent(req.Body)
	case oapi.WsEventDRAWSEND:
		return c.sendDrawSendEvent(req.Body)
	case oapi.WsEventANSWERREADY:
		return c.sendAnswerReadyEvent(req.Body)
	case oapi.WsEventANSWERCANCEL:
		return c.sendAnswerCancelEvent(req.Body)
	case oapi.WsEventANSWERSEND:
		return c.sendAnswerSendEvent(req.Body)
	case oapi.WsEventSHOWNEXT:
		return c.sendShowNextEvent(req.Body)
	case oapi.WsEventRETURNROOM:
		return c.sendReturnRoomEvent(req.Body)
	default:
		return errUnknownEventType
	}
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

// ODAI_READY
// お題の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) sendOdaiReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.server.room.Game.AddReady(c.userId)

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
		if err := c.server.sendOdaiInputEvent(c.server.room.Game.ReadyCount()); err != nil {
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

	c.server.room.Game.CancelReady(c.userId)

	if err := c.server.sendOdaiInputEvent(c.server.room.Game.ReadyCount()); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventODAIINPUT)
	}

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
					logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
				}
			case <-game.TimerStopChan:
			}
		}()
	}

	return nil
}

// DRAW_READY
// 絵が書き終わっていることを通知する (ルームの各員 -> サーバー)
func (c *Client) sendDrawReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.server.room.Game.AddReady(c.userId)

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
		if err := c.server.sendDrawInputEvent(c.server.room.Game.ReadyCount()); err != nil {
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

	c.server.room.Game.CancelReady(c.userId)

	if err := c.server.sendDrawInputEvent(c.server.room.Game.ReadyCount()); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventDRAWINPUT)
	}

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
						logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventDRAWFINISH).Error())
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

// ANSWER_READY
// 回答の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) sendAnswerReadyEvent(_ interface{}) error {
	if !c.server.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.server.room.Game.AddReady(c.userId)

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
		if err := c.server.sendAnswerInputEvent(c.server.room.Game.ReadyCount()); err != nil {
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

	c.server.room.Game.CancelReady(c.userId)

	if err := c.server.sendAnswerInputEvent(c.server.room.Game.ReadyCount()); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventANSWERINPUT)
	}

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
			logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventSHOWODAI))
		}
		game.NextShowPhase = model.GameShowPhaseCanvas
	case model.GameShowPhaseCanvas:
		if err := c.server.sendShowCanvasEvent(); err != nil {
			logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventSHOWCANVAS))
		}
		game.NextShowPhase = model.GameShowPhaseAnswer
	case model.GameShowPhaseAnswer:
		if err := c.server.sendShowAnswerEvent(); err != nil {
			logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventSHOWANSWER))
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
		logger.Echo.Error(c.server.sendEventErr(err, oapi.WsEventNEXTROOM))
	}

	return nil
}
