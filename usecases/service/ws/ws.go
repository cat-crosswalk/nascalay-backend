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
	hub             *Hub
	upgrader        websocket.Upgrader
	userIdToClients map[model.UserId]map[*Client]bool
}

func NewStreamer() Streamer {
	stream := &streamer{
		hub: NewHub(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		userIdToClients: make(map[model.UserId]map[*Client]bool),
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

	return nil
}

func (s *streamer) addNewClient(userId model.UserId, conn *websocket.Conn) *Client {
	cli := NewClient(userId, conn)
	s.hub.Register(cli)

	m, ok := s.userIdToClients[userId]
	if !ok {
		m = make(map[*Client]bool)
		s.userIdToClients[userId] = m
	}
	m[cli] = true

	return cli
}
