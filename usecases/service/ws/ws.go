package ws

import (
	"fmt"
	"net/http"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
<<<<<<< Updated upstream
=======
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/logger"
>>>>>>> Stashed changes
	"github.com/gorilla/websocket"
)

<<<<<<< Updated upstream
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
=======
type Hub struct {
	upgrader       websocket.Upgrader
	repo           repository.Repository
	userIdToClient map[model.UserId]*Client
	registerCh     chan *Client
	unregisterCh   chan *Client
	mux            sync.RWMutex
}

func InitHub(repo repository.Repository) *Hub {
	hub := &Hub{
>>>>>>> Stashed changes
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

<<<<<<< Updated upstream
func (s *streamer) addNewClient(userId model.UserId, conn *websocket.Conn) (*Client, error) {
	cli, err := NewClient(s.hub, userId, conn, s.logger)
=======
func (h *Hub) register(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	logger.Echo.Infof("new client(userId:%s) has registered", cli.userId.UUID().String())

	h.userIdToClient[cli.userId] = cli
}

func (h *Hub) unregister(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	logger.Echo.Infof("client(userId:%s) has unregistered", cli.userId.UUID().String())

	close(cli.send)
	delete(h.userIdToClient, cli.userId)
}

func (h *Hub) addNewClient(userId model.UserId, conn *websocket.Conn) (*Client, error) {
	cli, err := NewClient(h, userId, conn)
>>>>>>> Stashed changes
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
