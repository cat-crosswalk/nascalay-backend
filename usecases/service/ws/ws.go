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
		upgrader: websocket.Upgrader{},
	}
}

func (s *streamer) ServeWS(w http.ResponseWriter, r *http.Request, userId uuid.UUID) error {
	_, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	// TODO: かく
	return nil
}
