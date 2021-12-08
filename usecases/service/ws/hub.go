package ws

import (
	"github.com/21hack02win/nascalay-backend/model"
	"github.com/21hack02win/nascalay-backend/usecases/repository"
)

type Hub struct {
	repo            repository.Repository
	userIdToClients map[model.UserId]clientMap
	registerCh      chan *Client
	unregisterCh    chan *Client
}

type clientMap map[*Client]struct{}

func NewHub(repo repository.Repository) *Hub {
	return &Hub{
		repo:            repo,
		userIdToClients: make(map[model.UserId]clientMap),
		registerCh:      make(chan *Client),
		unregisterCh:    make(chan *Client),
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
			h.unregister(cli)
		}
	}
}

func (h *Hub) register(cli *Client) {
	h.userIdToClients[cli.userId] = clientMap{cli: {}}
}

func (h *Hub) unregister(cli *Client) {
	close(cli.send)
	delete(h.userIdToClients[cli.userId], cli)
}
