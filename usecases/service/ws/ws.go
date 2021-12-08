package ws

import (
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/gorilla/websocket"
)

type Streamer interface {
	Run()
	ServeWS(w http.ResponseWriter, r *http.Request, uid model.UserId) error
}

type streamer struct {
	hub      *Hub
	upgrader websocket.Upgrader
}

func NewStreamer(hub *Hub) Streamer {
	stream := &streamer{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	stream.Run()
	return stream
}

func (s *streamer) Run() {
	go s.hub.Run()
}

func (s *streamer) ServeWS(w http.ResponseWriter, r *http.Request, userId model.UserId) error {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	cli := s.addNewClient(userId, conn)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cli.writePump()
	go cli.readPump()

	cli.send <- []byte("Welcome to nascalay-backend!")

	return nil
}

func (s *streamer) addNewClient(userId model.UserId, conn *websocket.Conn) *Client {
	cli := NewClient(s.hub, userId, conn)
	s.hub.Register(cli)

	m, ok := s.hub.userIdToClients[userId]
	if !ok {
		m = make(clientMap)
		s.hub.userIdToClients[userId] = m
	}
	m[cli] = struct{}{}

	return cli
}
