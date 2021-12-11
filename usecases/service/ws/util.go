package ws

import (
	"log"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
)

// Exec `next` func for all clients in the room
func (c *Client) bloadcast(next func(c *Client)) {
	for _, m := range c.room.Members {
		cc, ok := c.hub.userIdToClient[m.Id]
		if !ok {
			continue
		}

		go next(cc)
	}
}

// Send message to a client
func (c *Client) sendMsg(msg []byte) {
	if c.send == nil {
		if c.userId == c.room.HostId {
			if err := c.sendChangeHostEvent(); err != nil {
				log.Println("failed to send change host event:", err.Error())
			}
		}
		c.hub.unregister(c)

		return
	}

	c.send <- msg
}

// Send message to all clients in the room
func (c *Client) sendMsgToEachClientInRoom(msg []byte) {
	c.bloadcast(func(cc *Client) {
		cc.sendMsg(msg)
	})
}

// Check if all members are ready
func (c *Client) allMembersAreReady() bool {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	r := c.room
	for _, m := range r.Members {
		if _, ok := c.hub.userIdToClient[m.Id]; !ok {
			continue
		}

		if _, ok := r.Game.Ready[m.Id]; !ok {
			return false
		}
	}

	return true
}

// Reset break timer
func (c *Client) resetBreakTimer() {
	// タイマーをリセットする
	// 15(分)後に次のゲームが始まらなければルームを削除する
	game := c.room.Game
	bt := game.BreakTimer

	if stopped := bt.Stop(); !stopped {
		go c.waitAndBreakRoom()
	}

	bt.Reset(time.Minute * 15)
	go c.waitAndBreakRoom()
}

// Wait for 15 minutes and break the room
func (c *Client) waitAndBreakRoom() {
	<-c.room.Game.BreakTimer.C
	if err := c.sendBreakRoomEvent(); err != nil {
		log.Println("failed to send BREAK_ROOM event:", err.Error())
	}
}

// Make the client ready for waiting
func (c *Client) AddReady(uid model.UserId) {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	c.room.Game.Ready[uid] = struct{}{}
}

// Cancel the client's ready state
func (c *Client) CancelReady(uid model.UserId) {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	delete(c.room.Game.Ready, uid)
}
