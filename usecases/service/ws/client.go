//nolint:errcheck
package ws

import (
	"fmt"
	"log"
	"time"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
)

type Client struct {
	hub    *Hub
	userId model.UserId
	room   *model.Room
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(hub *Hub, userId model.UserId, conn *websocket.Conn) (*Client, error) {
	room, err := hub.repo.GetRoomFromUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get room from userId: %w", err)
	}

	return &Client{
		hub:    hub,
		userId: userId,
		room:   room,
		conn:   conn,
		send:   make(chan []byte, 256),
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
				log.Printf("failed to create next writer: %v", err)
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				log.Println("failed to close writer:", err.Error())
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("failed to write message:", err.Error())
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
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		req := new(oapi.WsJSONRequestBody)
		if err := c.conn.ReadJSON(req); err != nil {
			if !websocket.IsCloseError(err) && !websocket.IsUnexpectedCloseError(err) {
				log.Println("websocket error occured:", err.Error())
			}
			break
		}

		if err := c.callEventHandler(req); err != nil {
			log.Println("websocket error occured:", err.Error())
			c.send <- []byte(err.Error())
			continue
		}
	}
}

func (c *Client) bloadcast(next func(c *Client)) {
	for _, m := range c.room.Members {
		c.hub.mux.Lock()
		defer c.hub.mux.Unlock()

		cc, ok := c.hub.userIdToClient[m.Id]
		if !ok {
			continue
		}

		next(cc)
	}
}

func (c *Client) sendMsg(msg []byte) {
	// TODO: unregisterできるようにする
	if c.send == nil {
		if c.userId == c.room.HostId {
			c.sendChangeHostEvent()
		}
		c.hub.unregister(c)

		return
	}

	c.send <- msg
}

func (c *Client) sendMsgToEachClientInRoom(msg []byte) {
	c.bloadcast(func(cc *Client) {
		c.sendMsg(msg)
	})
}

func (c *Client) allMembersAreReady() bool {
	r := c.room
	for _, m := range r.Members {
		c.hub.mux.Lock()
		defer c.hub.mux.Unlock()

		if _, ok := c.hub.userIdToClient[m.Id]; !ok {
			continue
		}

		if _, ok := r.Game.Ready[m.Id]; !ok {
			return false
		}
	}

	return true
}

func (c *Client) resetTimer() {
	// タイマーをリセットする
	// game.Timeout(分)後に次のゲームが始まらなければルームを削除する
	game := c.room.Game
	timer := game.Timer

	if stopped := timer.Stop(); !stopped {
		go c.waitAndBreakRoom()
	}

	timer.Reset(time.Minute * time.Duration(game.Timeout))
	go c.waitAndBreakRoom()
}

func (c *Client) waitAndBreakRoom() {
	<-c.room.Game.Timer.C
	if err := c.sendBreakRoomEvent(); err != nil {
		log.Println("failed to send BREAK_ROOM event:", err.Error())
	}
}
