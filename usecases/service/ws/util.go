package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
)

// Send message to a client
func (s *RoomServer) sendMsgTo(c *Client, msg *oapi.WsSendMessage) {
	// If the client is not connected, remove the client from the room
	// If the client is the host, change the host
	if c.send == nil {
		if c.userId == s.room.HostId {
			if err := s.sendChangeHostEvent(); err != nil {
				c.logger.Error(s.sendEventErr(err, oapi.WsEventCHANGEHOST))
			}
		}

		s.hub.unregister(c)

		return
	}

	c.send <- msg
}

// Send message to all clients in the room
func (s *RoomServer) sendMsgToEachClientInRoom(msg *oapi.WsSendMessage) {
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
func (s *RoomServer) allMembersAreReady() bool {
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
func (s *RoomServer) resetBreakTimer() {
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
func (s *RoomServer) waitAndBreakRoom() {
	<-s.room.Game.BreakTimer.C
	if err := s.sendBreakRoomEvent(); err != nil {
		s.logger.Error(s.sendEventErr(err, oapi.WsEventBREAKROOM))
	}
}

// Make the client ready for waiting
func (c *Client) AddReady(uid model.UserId) {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	c.server.room.Game.Ready[uid] = struct{}{}
}

// Cancel the client's ready state
func (c *Client) CancelReady(uid model.UserId) {
	c.hub.mux.Lock()
	defer c.hub.mux.Unlock()

	delete(c.server.room.Game.Ready, uid)
}

func (s *RoomServer) sendEventErr(err error, eventName oapi.WsEvent) error {
	return fmt.Errorf(
		"[ERROR] failed to send %s event (roomId:%s): %w",
		eventName,
		s.room.Id.String(),
		err,
	)
}
