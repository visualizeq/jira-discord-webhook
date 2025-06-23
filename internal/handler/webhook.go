package handler

import (
	"encoding/json"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"jira-discord-webhook/internal/discord"
	"jira-discord-webhook/internal/jira"
)

// WebhookHandler handles incoming Jira webhook requests and sends them to Discord.
func WebhookHandler(c *fiber.Ctx) error {
	// Debug log: raw payload received from Jira
	if ce := zap.L().Check(zap.DebugLevel, "JIRA payload"); ce != nil {
		ce.Write(zap.ByteString("payload", c.Body()))
	}

	var payload jira.Webhook
	if err := c.BodyParser(&payload); err != nil {
		zap.L().Error("failed to decode JIRA payload", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("bad request")
	}

	baseURL := os.Getenv("JIRA_BASE_URL")
	msg := jira.ToDiscordMessage(payload, baseURL)

	// Debug log: payload sent to Discord
	if ce := zap.L().Check(zap.DebugLevel, "Discord payload"); ce != nil {
		if b, err := json.Marshal(msg); err == nil {
			ce.Write(zap.ByteString("payload", b))
		}
	}

	if err := discord.SendFunc(msg); err != nil {
		zap.L().Error("failed to send to Discord", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("failed to send to Discord")
	}

	return c.SendStatus(fiber.StatusOK)
}
