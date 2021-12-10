//nolint:unused // TODO: 実装したら消す
package ws

import (
	"encoding/json"
	"fmt"
	"log"

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

// TODO: 実装する
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

	return nil
}

// TODO: 実装する
// ROOM_UPDATE_OPTION
// ゲームの設定を更新する (サーバー -> ルーム全員)
func (c *Client) sendRoomUpdateOptionEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

	return nil
}

// REQUEST_GAME_START
// ゲームを開始する (ホスト -> サーバー)
func (c *Client) receiveRequestGameStartEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
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

	// ODAIフェーズに移行
	c.room.Game.Status = model.GameStatusOdai

	go c.sendMsgToEachClientInRoom(buf)

	return nil
}

// ODAI_READY
// お題の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveOdaiReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusOdai) {
		return errWrongPhase
	}

	c.room.Game.AddReady(c.userId)

	if c.room.AllMembersAreReady() {
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

	c.room.Game.CancelReady(c.userId)

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

	go c.sendMsgToEachClientInRoom(buf)

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
	}

	c.bloadcast(func(cc *Client) {
		if err := cc.sendDrawStartEvent(); err != nil {
			log.Println("failed to send DRAW_START event:", err.Error())
		}
	})

	return nil
}

// DRAW_START
// キャンバス情報とお題を送信する (サーバー -> ルーム各員)
func (c *Client) sendDrawStartEvent() error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	var (
		game      = c.room.Game
		drawCount = game.DrawCount
		odai      *model.Odai
		drawer    *model.Drawer
	)

	random.SetupMemberRoles(game, c.room.Members)

	for _, v := range game.Odais {
		if v.DrawerSeq[drawCount].UserId == c.userId {
			odai = v
			drawer = &v.DrawerSeq[drawCount]
			break
		}
	}

	if odai == nil {
		return errNotFound
	}

	if drawer == nil {
		return errUnAuthorized
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventDRAWSTART,
			Body: oapi.WsDrawStartEventBody{
				AllDrawPhaseNum: game.AllDrawPhase(),
				Canvas: oapi.Canvas{
					AreaId:    drawer.AreaId.Int(),
					BoardName: game.Canvas.BoardName,
				},
				DrawPhaseNum: game.DrawCount.Int(),
				Img:          string(odai.Img),
				Odai:         odai.Title.String(),
				TimeLimit:    int(game.TimeLimit),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.send <- buf

	return nil
}

// DRAW_READY
// 絵が書き終わっていることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveDrawReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	c.room.Game.AddReady(c.userId)

	if c.room.AllMembersAreReady() {
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

	c.room.Game.CancelReady(c.userId)

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

	go c.sendMsgToEachClientInRoom(buf)

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
			if len(v.Img) == 0 {
				newImg, err := canvas.MergeImage(model.Img(v.Img), model.Img(e.Img))
				if err != nil {
					v.Img = model.Img(e.Img)
					return fmt.Errorf("failed to merge image: %w", err)
				}
				v.Img = newImg
			} else {
				v.Img = model.Img(e.Img)
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
		if game.DrawCount.Int()+1 < game.AllDrawPhase() {
			game.DrawCount++

			game.ResetImgUpdated()

			c.bloadcast(func(cc *Client) {
				if err := cc.sendDrawStartEvent(); err != nil {
					log.Println("failed to send DRAW_START event:", err)
				}
			})
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

	for _, v := range c.room.Game.Odais {
		ac, ok := c.hub.userIdToClient[v.AnswererId]
		if !ok {
			return errNotFound
		}

		buf, err := json.Marshal(
			&oapi.WsJSONBody{
				Type: oapi.WsEventANSWERSTART,
				Body: oapi.WsAnswerStartEventBody{
					Img:       string(v.Img),
					TimeLimit: int(c.room.Game.TimeLimit),
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to encode as JSON: %w", err)
		}

		ac.send <- buf
	}

	return nil
}

// ANSWER_READY
// 回答の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveAnswerReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	c.room.Game.AddReady(c.userId)

	if c.room.AllMembersAreReady() {
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

	c.room.Game.CancelReady(c.userId)

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

	go c.sendMsgToEachClientInRoom(buf)

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
			v.Answer = model.OdaiAnswer(e.Answer)
			break
		}
	}

	// 全員の回答が送信されたらSHOWフェーズに移行
	allAnswerReceived := true
	for _, v := range game.Odais {
		if v.Answer == "" {
			allAnswerReceived = false
			break
		}
	}

	if allAnswerReceived {
		game.Status = model.GameStatusShow

		c.bloadcast(func(cc *Client) {
			if err := cc.sendShowStartEvent(); err != nil {
				log.Println("failed to send SHOW_START event:", err)
			}
		})
	}

	return nil
}

// SHOW_START
// 結果表示フェーズが始まったことを通知する (サーバー -> ルーム全員)
func (c *Client) sendShowStartEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWSTART,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.send <- buf

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
		if err := c.sendShowOdaiEvent(); err != nil {
			return fmt.Errorf("failed to send SHOW_ODAI event: %w", err)
		}
	case model.GameShowPhaseCanvas:
		if err := c.sendShowCanvasEvent(); err != nil {
			return fmt.Errorf("failed to send SHOW_CANVAS event: %w", err)
		}
	case model.GameShowPhaseAnswer:
		if err := c.sendShowAnswerEvent(); err != nil {
			return fmt.Errorf("failed to send SHOW_ANSWER event: %w", err)
		}
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
		game      = c.room.Game
		showCount = game.ShowCount
	)
	if len(game.Odais) < showCount.Int()+1 {
		return errNotFound
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWODAI,
			Body: &oapi.WsShowOdaiEventBody{
				Next: oapi.WsNextShowStatus("canvas"),
				Odai: game.Odais[showCount.Int()].Title.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.send <- buf

	return nil
}

// SHOW_CANVAS
// 次のキャンバスを受信する (サーバー -> ルーム全員)
func (c *Client) sendShowCanvasEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}
	var (
		game      = c.room.Game
		showCount = game.ShowCount
	)
	if len(game.Odais) < showCount.Int()+1 {
		return errNotFound
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWCANVAS,
			Body: &oapi.WsShowCanvasEventBody{
				Next: oapi.WsNextShowStatus("answer"),
				Img:  game.Odais[showCount.Int()].Img.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.send <- buf

	return nil
}

// SHOW_ANSWER
// 最後の回答を受信する (サーバー -> ルーム全員)
func (c *Client) sendShowAnswerEvent() error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}
	var (
		game      = c.room.Game
		showCount = game.ShowCount
	)
	if len(game.Odais) < showCount.Int()+1 {
		return errNotFound
	}

	next := oapi.WsNextShowStatus("odai")
	if len(game.Odais) == showCount.Int()+1 {
		next = oapi.WsNextShowStatus("end")
	}

	buf, err := json.Marshal(
		&oapi.WsJSONBody{
			Type: oapi.WsEventSHOWANSWER,
			Body: &oapi.WsShowAnswerEventBody{
				Next: next,
				Answer: game.Odais[showCount.Int()].Answer.String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to encode as JSON: %w", err)
	}

	c.send <- buf

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

	c.bloadcast(func(cc *Client) {
		if err := cc.sendNextRoomEvent(); err != nil {
			log.Println("failed to send NEXT_ROOM event:", err)
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

	c.send <- buf

	return nil
}

// TODO: 実装する
// CHANGE_HOST
// ホストが落ちた時に飛んできて，ホスト役を変更する (サーバー -> ルーム全員)
func (c *Client) sendChangeHostEvent(body interface{}) error {
	return nil
}

// TODO: 実装する
// BREAK_ROOM
// 部屋が破壊されたときに通知する (サーバー -> ルーム全員)
// 部屋が立ってからゲーム開始まで15分以上経過している場合，部屋を閉じる
// このタイミングでサーバーは保持しているルームに関わる全データを削除
func (c *Client) sendBreakRoomEvent(body interface{}) error {
	return nil
}
