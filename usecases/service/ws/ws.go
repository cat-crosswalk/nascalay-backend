package ws

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Streamer interface {
	// Run()
	ServeWS(w http.ResponseWriter, r *http.Request, userId uuid.UUID) error
}

type streamer struct {
	// hub *Hub
	upgrader websocket.Upgrader
}

func NewStreamer() Streamer {
	return &streamer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *streamer) ServeWS(w http.ResponseWriter, r *http.Request, userId uuid.UUID) error {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	// TODO: かく
	return nil
}
