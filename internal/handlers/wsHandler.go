package handlers

import (
	"encoding/json"
	"fiber-go/internal/middleware"
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

	userID, ok := c.Locals(middleware.UserIDKey).(int)
	if !ok {
		return fiber.ErrUnauthorized
	}

	handler := websocket.New(func(conn *websocket.Conn) {

		client := h.hub.Register(userID, conn)

		defer h.hub.Unregister(client)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var m models.WSMessage
			if err := json.Unmarshal(msg, &m); err != nil {
				continue
			}

			switch m.Type {

			case "chat":
				h.hub.SendToUser(m.To, models.WSMessage{
					Type: "chat",
					From: userID,
					Body: m.Body,
				})
			}
		}
	})

	return handler(c)
}
