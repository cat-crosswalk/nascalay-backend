package ws

type Hub struct {
	// repo         repository.Repository
	clients      map[*Client]struct{}
	registerCh   chan *Client
	unregisterCh chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]struct{}),
		registerCh:   make(chan *Client),
		unregisterCh: make(chan *Client),
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
	h.clients[cli] = struct{}{}
}

func (h *Hub) unregister(cli *Client) {
	delete(h.clients, cli)
}
