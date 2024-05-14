package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
	"github.com/21hack02win/nascalay-backend/util/logger"
	"github.com/gorilla/websocket"
)

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
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		repo:           repo,
		userIdToClient: make(map[model.UserId]*Client),
		registerCh:     make(chan *Client),
		unregisterCh:   make(chan *Client),
		mux:            sync.RWMutex{},
	}

	go hub.run()

	return hub
}

func (h *Hub) run() {
	for {
		select {
		case cli := <-h.registerCh:
			h.register(cli)
		case cli := <-h.unregisterCh:
			if _, ok := h.userIdToClient[cli.userId]; ok {
				h.unregister(cli)
			}
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userId model.UserId) error {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade the HTTP server connection to the WebSocket protocol: %w", err)
	}

	cli, err := h.addNewClient(userId, conn)
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

func (h *Hub) NotifyOfNewRoomMember(room *model.Room) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	c, ok := h.userIdToClient[room.HostId]
	if !ok {
		return errNotFound
	}

	if err := c.server.sendRoomNewMemberEvent(room); err != nil {
		return c.server.sendEventErr(err, oapi.WsEventROOMNEWMEMBER)
	}

	return nil
}

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
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %w", err)
	}

	h.registerCh <- cli

	h.mux.Lock()
	defer h.mux.Unlock()

	c, ok := h.userIdToClient[userId]
	if !ok {
		h.userIdToClient[userId] = c
	}

	return cli, nil
}
