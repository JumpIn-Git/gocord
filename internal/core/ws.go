package core

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Event struct {
	Type     string `json:"type"`
	ServerID int64  `json:"server_id"`
	Payload  any    `json:"payload"`
}

type Client struct {
	Conn      *websocket.Conn
	UserID    int64
	ServerIDs []int64
	Send      chan Event
}

func (c *Client) WritePump() {
	for {
		select {
		case event := <-c.Send:
			c.Conn.WriteJSON(event)
		default:
			continue
		}
	}
}

type Hub struct {
	Broadcast chan Event
	Servers   map[int64]map[*Client]struct{}
	Login     chan *Client
	Logout    chan *Client
}

var hub = Hub{
	Broadcast: make(chan Event),
	Servers:   make(map[int64]map[*Client]struct{}),
	Login:     make(chan *Client),
	Logout:    make(chan *Client),
}

func GetHub() *Hub {
	return &hub
}

func (h *Hub) Run() {
	for {
		select {
		case event := <-h.Broadcast:
			if clients, ok := h.Servers[event.ServerID]; ok {
				for client := range clients {
					select {
					case client.Send <- event:
					default:
						h.removeClient(client)
						close(client.Send)
					}
				}
			}
		case client := <-h.Login:
			for _, id := range client.ServerIDs {
				if _, ok := h.Servers[id]; !ok {
					h.Servers[id] = make(map[*Client]struct{})
				}
				h.Servers[id][client] = struct{}{}
			}
		case client := <-h.Logout:
			h.removeClient(client)
			close(client.Send)
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	for _, id := range client.ServerIDs {
		if clients, ok := h.Servers[id]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.Servers, id)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) HandleWebSocket(c echo.Context) error {
	UserID := c.Get("user_id").(int64)

	serverIDs, err := s.Q.GetUserServersIDs(c.Request().Context(), UserID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load servers")
	}
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to upgrade connection")
	}
	client := &Client{
		Conn:      conn,
		UserID:    UserID,
		ServerIDs: serverIDs,
		Send:      make(chan Event),
	}
	hub.Login <- client
	go client.WritePump()
	return nil
}
