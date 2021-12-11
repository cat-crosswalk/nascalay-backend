package ws

import (
	"sync"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
)

type Hub struct {
	repo           repository.Repository
	userIdToClient map[model.UserId]*Client
	registerCh     chan *Client
	unregisterCh   chan *Client
	mux            sync.RWMutex
}

func NewHub(repo repository.Repository) *Hub {
	return &Hub{
		repo:           repo,
		userIdToClient: make(map[model.UserId]*Client),
		registerCh:     make(chan *Client),
		unregisterCh:   make(chan *Client),
		mux:            sync.RWMutex{},
	}
}

func (h *Hub) Register(client *Client) {
	h.registerCh <- client
}
func (h *Hub) Unregister(client *Client) {
	h.unregisterCh <- client
}

func (h *Hub) Run() {
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

func (h *Hub) register(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.userIdToClient[cli.userId] = cli
}

func (h *Hub) unregister(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	close(cli.send)
	delete(h.userIdToClient, cli.userId)
}
