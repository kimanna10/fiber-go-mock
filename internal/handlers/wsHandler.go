package handlers

import (
	"encoding/json"
	"fiber-go/internal/models"
	"fiber-go/internal/services"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

type WSHandler struct {
	hub *services.WSHub
}

func NewWSHandler(h *services.WSHub) *WSHandler {
	return &WSHandler{hub: h}
}

func (h *WSHandler) HandleWS(c fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}

	userID, ok := c.Locals("user_id").(int)
	if !ok {
		return fiber.ErrUnauthorized
	}
	handler := websocket.New(func(conn *websocket.Conn) {
		client := h.hub.Register(userID, conn)
		defer h.hub.Unregister(client)
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if messageType == websocket.TextMessage {
				var event models.WSEvent
				if err := json.Unmarshal(p, &event); err != nil {
					continue
				}
				switch event.Type {
				case models.EventMessage:
					var msg models.WSMessage
					if err := json.Unmarshal(event.Data, &msg); err != nil {
						continue
					}
					msg.SenderID = userID
					h.hub.BroadcastToChat(msg)

				case models.EventTyping:
					var typing models.TypingEvent
					if err := json.Unmarshal(event.Data, &typing); err != nil {
						continue
					}
					typing.UserID = userID
					h.hub.BroadcastTyping(typing)
				}
			}

		}
	})

	return handler(c)
}
