package ws

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gorilla/websocket"
)

type Client struct {
	userId model.UserId
	conn   *websocket.Conn
}

func NewClient(userId model.UserId, conn *websocket.Conn) *Client {
	return &Client{
		userId: userId,
		conn:   conn,
	}
}
