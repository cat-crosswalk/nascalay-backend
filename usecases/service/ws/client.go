//nolint:errcheck
package ws

import (
	"encoding/json"
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
	maxMessageSize = 300000
)

type Client struct {
	hub    *Hub
	userId model.UserId
	room   *model.Room
	conn   *websocket.Conn
	send   chan *oapi.WsSendMessage
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
				log.Println("failed to create next writer:", err.Error())
				return
			}

			buf, err := json.Marshal(message)
			if err != nil {
				log.Println("failed to encode as JSON:", err.Error())
				return
			}

			w.Write(buf)

			// Add queued chat messages to the current websocket message.
			for i := 0; i < len(c.send); i++ {
				buf, err = json.Marshal(<-c.send)
				if err != nil {
					log.Println("failed to encode as JSON:", err.Error())
					return
				}

				w.Write(buf)
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
