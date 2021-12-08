//nolint:unused // TODO: 実装したら消す
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

// ROOM_NEW_MEMBER
// 部屋に追加のメンバーが来たことを通知する (サーバー -> ルーム全員)
func (c *Client) sendRoomNewMemberEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusRoom) {
		return errWrongPhase
	}

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
		return err
	}

	return nil
}

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

	c.room.Game.Status = model.GameStatusOdai

	if err := c.sendGameStartEvent(); err != nil {
		return err
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

// ODAI_READY
// お題の入力が完了していることを通知する (ルームの各員 -> サーバー)
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
		return err
	}

	go c.sendToEachClientInRoom(buf)

	return nil
}

// ODAI_SEND
// お題を送信する (ルームの各員 -> サーバー)
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

// DRAW_START
// キャンバス情報とお題を送信する (サーバー -> ルーム各員)
func (c *Client) sendDrawStartEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

// DRAW_READY
// 絵が書き終わっていることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveDrawReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

// DRAW_CANCEL
// 絵が書き終わっている通知を解除する (ルームの各員 -> サーバー)
func (c *Client) receiveDrawCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

// DRAW_FINISH
// 全員が絵を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは絵を送信する
func (c *Client) sendDrawFinishEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusDraw) {
		return errWrongPhase
	}

	return nil
}

// DRAW_SEND
// 絵を送信する (ルームの各員 -> サーバー)
// -> (DRAWフェーズが終わってなかったら) また，DRAW_START が飛んでくる
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

// ANSWER_START
// 絵が飛んできて，回答する (サーバー -> ルーム各員)
func (c *Client) sendAnswerStartEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// ANSWER_READY
// 回答の入力が完了していることを通知する (ルームの各員 -> サーバー)
func (c *Client) receiveAnswerReadyEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// ANSWER_CANCEL
// 回答の入力の完了を解除する (ルームの各員 -> サーバー)
func (c *Client) receiveAnswerCancelEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// ANSWER_FINISH
// 全員が回答の入力を完了したことor制限時間が来たことを通知する (サーバー -> ルーム全員)
// クライアントは回答を送信する
func (c *Client) sendAnswerFinishEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// ANSWER_SEND
// 回答を送信する (ルームの各員 -> サーバー)
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

// SHOW_START
// 結果表示フェーズが始まったことを通知する (サーバー -> ルーム全員)
func (c *Client) sendShowStartEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// SHOW_NEXT
// つぎの結果表示を要求する (ホスト -> サーバー)
func (c *Client) receiveShowNextEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// SHOW_ODAI
// 最初のお題を受信する (サーバー -> ルーム全員)
func (c *Client) sendShowOdaiEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// SHOW_CANVAS
// 次のキャンバスを受信する (サーバー -> ルーム全員)
func (c *Client) sendShowCanvasEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// SHOW_ANSWER
// 最後の回答を受信する (サーバー -> ルーム全員)
func (c *Client) sendShowAnswerEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// RETURN_ROOM
// ルーム(新規加入待機状態) に戻る (ホスト -> サーバー)
// このタイミングでサーバーは保持しているゲームデータを削除
func (c *Client) receiveReturnRoomEvent(_ interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusAnswer) {
		return errWrongPhase
	}

	return nil
}

// NEXT_ROOM
// ルームの表示に遷移する (サーバー -> ルーム全員)
func (c *Client) sendNextRoomEvent(body interface{}) error {
	if !c.room.GameStatusIs(model.GameStatusShow) {
		return errWrongPhase
	}

	return nil
}

// CHANGE_HOST
// ホストが落ちた時に飛んできて，ホスト役を変更する (サーバー -> ルーム全員)
func (c *Client) sendChangeHostEvent(body interface{}) error {
	return nil
}

// BREAK_ROOM
// 部屋が破壊されたときに通知する (サーバー -> ルーム全員)
// 部屋が立ってからゲーム開始まで15分以上経過している場合，部屋を閉じる
// このタイミングでサーバーは保持しているルームに関わる全データを削除
func (c *Client) sendBreakRoomEvent(body interface{}) error {
	return nil
}
