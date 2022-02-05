package ws

import (
	"sync"

	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
)

type Hub struct {
	repo           repository.Repository
	userIdToClient map[model.UserId]*Client
	mux            sync.RWMutex
}

func NewHub(repo repository.Repository) *Hub {
	return &Hub{
		repo:           repo,
		userIdToClient: make(map[model.UserId]*Client),
		mux:            sync.RWMutex{},
	}
}

func (h *Hub) register(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	cli.logger.Infof("new client(userId:%s) has registered", cli.userId.UUID().String())

	h.userIdToClient[cli.userId] = cli
}

func (h *Hub) unregister(cli *Client) {
	h.mux.Lock()
	defer h.mux.Unlock()

	cli.logger.Infof("client(userId:%s) has unregistered", cli.userId.UUID().String())

	close(cli.send)
	delete(h.userIdToClient, cli.userId)
}
