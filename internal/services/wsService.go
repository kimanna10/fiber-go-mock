package services

import (
	"context"
	"encoding/json"
	"fiber-go/internal/cache"
	"fiber-go/internal/models"
	"sync"

	"github.com/gofiber/contrib/v3/websocket"
)

type Client struct {
	UserID int             // id пользователя
	Conn   *websocket.Conn // активное TCP-соединение с данным пользователем
}

type WSHub struct {
	mu      sync.RWMutex
	clients map[int]map[*Client]bool // мап активных девайсов
	cache   cache.Cache
	keys    *cache.KeyBuilder
}

func NewWSHub(cache cache.Cache, keys *cache.KeyBuilder) *WSHub {
	return &WSHub{
		clients: make(map[int]map[*Client]bool),
		cache:   cache,
		keys:    keys,
	}
}

func (h *WSHub) Register(userID int, conn *websocket.Conn) *Client {
	client := &Client{
		UserID: userID,
		Conn:   conn,
	}

	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]bool)
	}
	h.clients[userID][client] = true
	h.mu.Unlock()

	ctx := context.Background()
	key := h.keys.UserStatus(userID)
	_ = h.cache.Set(ctx, key, []byte("online"), 0)

	return client
}

func (h *WSHub) Unregister(client *Client) {
	h.mu.Lock()

	if _, ok := h.clients[client.UserID]; ok {
		delete(h.clients[client.UserID], client)

		if len(h.clients[client.UserID]) == 0 {
			delete(h.clients, client.UserID)
		}
	}
	h.mu.Unlock()

	client.Conn.Close()

	ctx := context.Background()
	key := h.keys.UserStatus(client.UserID)
	_ = h.cache.Set(ctx, key, []byte("offline"), 0)
}

func (h *WSHub) SendToUser(toUserID int, msg models.WSMessage) {
	h.mu.RLock()
	clients, ok := h.clients[toUserID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, _ := json.Marshal(msg)

	for client := range clients {
		_ = client.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
