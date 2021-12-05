package ws

import "github.com/21hack02win/nascalay-backend/model"

type Hub struct {
	clientsPerRoom map[model.RoomId]map[*Client]bool
	registerCh     chan *Client
	unregisterCh   chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clientsPerRoom: make(map[model.RoomId]map[*Client]bool),
		registerCh:     make(chan *Client),
		unregisterCh:   make(chan *Client),
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
	_ = cli.userId
	roomId := model.RoomId("test") // TODO: userIdからroomIdを取得する
	if _, ok := h.clientsPerRoom[roomId]; ok {
		h.clientsPerRoom[roomId][cli] = true
		return
	}
	h.clientsPerRoom[roomId] = map[*Client]bool{cli: true}
}

func (h *Hub) unregister(cli *Client) {
	_ = cli.userId
	roomId := model.RoomId("test") // TODO: userIdからroomIdを取得する
	delete(h.clientsPerRoom[roomId], cli)
}
