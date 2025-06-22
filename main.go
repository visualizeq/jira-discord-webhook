package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"jira-discord-webhook/internal/discord"
	"jira-discord-webhook/internal/jira"
)

func webhookHandler(c *fiber.Ctx) error {
	var payload jira.Webhook
	if err := c.BodyParser(&payload); err != nil {
		log.Println("failed to decode jira payload:", err)
		return c.Status(fiber.StatusBadRequest).SendString("bad request")
	}

	baseURL := os.Getenv("JIRA_BASE_URL")
	msg := jira.ToDiscordMessage(payload, baseURL)

	if err := discord.SendFunc(msg); err != nil {
		log.Println("failed to send to discord:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("failed to send to discord")
	}

	return c.SendStatus(fiber.StatusOK)
}

func main() {
	app := fiber.New()
	app.Post("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on :" + port)
	log.Fatal(app.Listen(":" + port))
}
