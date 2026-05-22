package services

import (
	"context"
	"encoding/json"
	"fiber-go/internal/cache"
	"fiber-go/internal/models"
	"fiber-go/internal/repository"
	"sync"

	"github.com/gofiber/contrib/v3/websocket"
)

type Client struct {
	UserID int             // id пользователя
	Conn   *websocket.Conn // активное TCP-соединение с данным пользователем
	mu     sync.Mutex      // Защищает конкретное соединение от одновременной записи нескольких горутин
}

type WSHub struct {
	mu          sync.RWMutex
	clients     map[int]map[*Client]bool // мап активных девайсов
	cache       cache.Cache
	keys        *cache.KeyBuilder
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
}

func NewWSHub(cache cache.Cache, keys *cache.KeyBuilder, messageRepo repository.MessageRepository, chatRepo repository.ChatRepository) *WSHub {
	return &WSHub{
		clients:     make(map[int]map[*Client]bool),
		cache:       cache,
		keys:        keys,
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
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

	go func() {
		ctx := context.Background()
		key := h.keys.UserStatus(userID)
		_ = h.cache.Set(ctx, key, []byte("online"), 0)
	}()
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
	go func() {
		ctx := context.Background()
		key := h.keys.UserStatus(client.UserID)
		_ = h.cache.Set(ctx, key, []byte("offline"), 0)
	}()
}

func (h *WSHub) BroadcastToChat(msg models.WSMessage) {
	go func(m models.WSMessage) {
		ctx := context.Background()
		if err := h.messageRepo.Create(ctx, &m); err != nil {
			return
		}
		h.dispatch(ctx, m)
	}(msg)
}

func (h *WSHub) dispatch(ctx context.Context, msg models.WSMessage) {
	chatInfo, err := h.chatRepo.GetChatInfo(ctx, msg.ChatID)
	if err != nil {
		return
	}

	data := buildMessageData(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()

	switch chatInfo.Type {
	case models.ChatPrivate, models.ChatGroup:
		for _, userId := range chatInfo.Members {
			if userId == msg.SenderID {
				continue
			}

			if devices, online := h.clients[userId]; online {
				for client := range devices {
					// message
					_ = client.Send(websocket.TextMessage, data)
				}
			}
		}

	case models.ChatBroadcast:
		for userId, devices := range h.clients {
			if userId == msg.SenderID {
				continue
			}
			for client := range devices {
				// message
				_ = client.Send(websocket.TextMessage, data)
			}
		}
	}

}

// Безопасный метод отправки, чтобы избежать паники concurrent write
func (c *Client) Send(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteMessage(messageType, data)
}

func (h *WSHub) BroadcastTyping(event models.TypingEvent) {
	// Запускаем в фоне, чтобы не блокировать хэндлер
	go func(ev models.TypingEvent) {
		ctx := context.Background()
		// Узнаем, кто участник этого чата
		chatInfo, err := h.chatRepo.GetChatInfo(ctx, ev.ChatID)
		if err != nil {
			return
		}

		payload, _ := json.Marshal(ev)
		response := models.WSEvent{
			Type: models.EventTyping,
			Data: payload,
		}
		data, _ := json.Marshal(response)

		h.mu.RLock()
		defer h.mu.RUnlock()

		// Рассылаем статус только участникам ЭТОГО чата
		for _, userId := range chatInfo.Members {
			if userId == ev.UserID {
				continue // Самим себе не шлем
			}
			if devices, online := h.clients[userId]; online {
				for client := range devices {
					_ = client.Send(websocket.TextMessage, data)
				}
			}
		}
	}(event)
}

func buildMessageData(msg models.WSMessage) []byte {
	payload, _ := json.Marshal(msg)
	response := models.WSEvent{
		Type: models.EventMessage,
		Data: payload,
	}
	data, _ := json.Marshal(response)
	return data
}

// func (h *WSHub) SendToUser(toUserID int, msg models.WSMessage) {
// 	h.mu.RLock()
// 	clients, ok := h.clients[toUserID]
// 	h.mu.RUnlock()

// 	if !ok {
// 		return
// 	}

// 	data, _ := json.Marshal(msg)

// 	for client := range clients {
// 		_ = client.Conn.WriteMessage(websocket.TextMessage, data)
// 	}
// }
