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
	Conn   *websocket.Conn
	UserID int64
	Send   chan Event
}

type LoginRequest struct {
	Client    *Client
	ServerIDs []int64
}

func (c *Client) WritePump() {
	for event := range c.Send {
		c.Conn.WriteJSON(event)
	}
}

type UserJoined struct {
	UserID int64
	Server int64
}

type UserLeft struct {
	UserID int64
	Server int64
}

type OnlineUser struct {
	Servers map[int64]struct{}
	Clients map[*Client]struct{}
}

// (using maps for 15+ items is faster than a slice, stuct{}{} is 0 bytes)
type Hub struct {
	Broadcast chan Event
	Servers   map[int64]map[int64]struct{} // server ID -> online user IDs
	Users     map[int64]*OnlineUser        // user ID -> active WebSocket sessions
	// UserServers map[int64]map[int64]struct{}   // user ID -> server IDs they are member of
	Login  chan LoginRequest
	Logout chan *Client
	Joined chan UserJoined
	Left   chan UserLeft
}

func NewHub() *Hub {
	return &Hub{
		Broadcast: make(chan Event),
		Servers:   make(map[int64]map[int64]struct{}),
		Users:     make(map[int64]*OnlineUser),
		// UserServers: make(map[int64]map[int64]struct{}),
		Login:  make(chan LoginRequest),
		Logout: make(chan *Client),
		Joined: make(chan UserJoined),
		Left:   make(chan UserLeft),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case event := <-h.Broadcast:
			if userIDs, ok := h.Servers[event.ServerID]; ok {
				for userID := range userIDs {
					if user, ok := h.Users[userID]; ok {
						for client := range user.Clients {
							select {
							case client.Send <- event:
							default:
								h.removeClient(client)
								close(client.Send)
							}
						}
					}
				}
			}
		case req := <-h.Login:
			if _, ok := h.Users[req.Client.UserID]; !ok {
				h.Users[req.Client.UserID] = &OnlineUser{
					Servers: make(map[int64]struct{}),
					Clients: make(map[*Client]struct{}),
				}
				for _, id := range req.ServerIDs {
					h.Users[req.Client.UserID].Servers[id] = struct{}{}
					if _, ok := h.Servers[id]; !ok {
						h.Servers[id] = make(map[int64]struct{})
					}
					h.Servers[id][req.Client.UserID] = struct{}{}
				}
			}
			h.Users[req.Client.UserID].Clients[req.Client] = struct{}{}
		case client := <-h.Logout:
			h.removeClient(client)
			close(client.Send)
		case joined := <-h.Joined:
			if _, ok := h.Servers[joined.Server]; !ok {
				h.Servers[joined.Server] = make(map[int64]struct{})
			}
			h.Servers[joined.Server][joined.UserID] = struct{}{}
			if user, ok := h.Users[joined.UserID]; ok {
				user.Servers[joined.Server] = struct{}{}
			}
		case left := <-h.Left:
			if users, ok := h.Servers[left.Server]; ok {
				delete(users, left.UserID)
				if len(users) == 0 {
					delete(h.Servers, left.Server)
				}
			}
			if user, ok := h.Users[left.UserID]; ok {
				delete(user.Servers, left.Server)
			}
		}

	}
}

func (h *Hub) removeClient(client *Client) {
	if user, ok := h.Users[client.UserID]; ok {
		delete(user.Clients, client)
		if len(user.Clients) == 0 { // last logged in session of account
			delete(h.Users, client.UserID)
			for serverID := range user.Servers {
				if users, ok := h.Servers[serverID]; ok {
					delete(users, client.UserID)
					if len(users) == 0 {
						delete(h.Servers, serverID)
					}
				}
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
		Conn:   conn,
		UserID: UserID,
		Send:   make(chan Event, 64),
	}
	s.Hub.Login <- LoginRequest{Client: client, ServerIDs: serverIDs}
	go client.WritePump()
	return nil
}
