package ws

import (
	"fmt"
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Streamer interface {
	Run()
	ServeWS(w http.ResponseWriter, r *http.Request, uid model.UserId) error
	NotifyOfNewRoomMember(room *model.Room) error
}

type streamer struct {
	hub      *Hub
	upgrader websocket.Upgrader
	logger   echo.Logger
}

func NewStreamer(hub *Hub, logger echo.Logger) Streamer {
	stream := &streamer{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		logger: logger,
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
		return fmt.Errorf("failed to upgrade the HTTP server connection to the WebSocket protocol: %w", err)
	}

	cli, err := s.addNewClient(userId, conn)
	if err != nil {
		return fmt.Errorf("failed to add new client: %w", err)
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cli.writePump()
	go cli.readPump()

	cli.send <- &oapi.WsSendMessage{
		Type: oapi.WsEventWELCOMENEWCLIENT,
		Body: oapi.WsWelcomeNewClientBody{
			Content: "Welcome to nascalay-backend!",
		},
	}

	return nil
}

func (s *streamer) NotifyOfNewRoomMember(room *model.Room) error {
	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	c, ok := s.hub.userIdToClient[room.HostId]
	if !ok {
		return errNotFound
	}

	if err := c.server.sendRoomNewMemberEvent(room); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventROOMNEWMEMBER)
	}

	return nil
}

func (s *streamer) addNewClient(userId model.UserId, conn *websocket.Conn) (*Client, error) {
	cli, err := NewClient(s.hub, userId, conn, s.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	s.hub.Register(cli)

	s.hub.mux.Lock()
	defer s.hub.mux.Unlock()

	c, ok := s.hub.userIdToClient[userId]
	if !ok {
		s.hub.userIdToClient[userId] = c
	}

	return cli, nil
}
