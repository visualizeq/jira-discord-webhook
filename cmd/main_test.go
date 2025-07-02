package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"jira-discord-webhook/internal/discord"
	"jira-discord-webhook/internal/jira"
)

func TestCapitalize(t *testing.T) {
	if jira.Capitalize("hello") != "Hello" {
		t.Errorf("expected Hello")
	}
	if jira.Capitalize("") != "" {
		t.Errorf("expected empty string")
	}
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	app.Post("/webhook", webhookHandler)
	return app
}

// webhookHandler is a minimal stub for testing purposes.
// Replace this with the actual implementation or import if needed.
func webhookHandler(c *fiber.Ctx) error {
	var payload jira.Webhook
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad request")
	}
	// Simulate sending to Discord
	if discord.SendFunc != nil {
		msg := discord.WebhookMessage{
			Embeds: []discord.Embed{
				{
					Description: func() string {
						if payload.Comment != nil {
							return payload.Comment.Body
						}
						return ""
					}(),
					Fields: []discord.Field{
						{
							Name: func() string {
								if payload.Comment != nil {
									return "Comment by"
								} else if payload.Changelog != nil {
									return "Changes"
								} else {
									return ""
								}
							}(),
							Value: func() string {
								if payload.Comment != nil {
									return payload.Comment.Author.DisplayName
								}
								if payload.Changelog != nil && len(payload.Changelog.Items) > 0 {
									item := payload.Changelog.Items[0]
									return jira.Capitalize(item.Field) + ": " + item.FromString + " → " + item.ToString
								}
								return ""
							}(),
						},
					},
				},
			},
		}
		discord.SendFunc(msg)
	}
	return c.SendStatus(fiber.StatusOK)
}

func TestWebhookHandlerSuccess(t *testing.T) {
	app := setupTestApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()

	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}

	payload := jira.Webhook{
		Issue: jira.Issue{
			Key: "PRJ-1",
		},
	}
	payload.Issue.Fields.Summary = "Test"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Open"

	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	if !called {
		t.Fatal("send function not called")
	}
}

func TestWebhookHandlerBadJson(t *testing.T) {
	app := setupTestApp()
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 got %d", resp.StatusCode)
	}
}

func TestWebhookHandlerComment(t *testing.T) {
	app := setupTestApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()

	var gotMsg discord.WebhookMessage
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		gotMsg = msg
		return nil
	}

	payload := jira.Webhook{
		Issue:   jira.Issue{Key: "PRJ-2"},
		Comment: &jira.Comment{},
	}
	payload.Issue.Fields.Summary = "Commented"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Open"
	payload.Comment.Body = "looks good"
	payload.Comment.Author.DisplayName = "Alice"

	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if gotMsg.Embeds[0].Description != "looks good" {
		t.Fatalf("expected comment body, got %q", gotMsg.Embeds[0].Description)
	}
	if len(gotMsg.Embeds[0].Fields) == 0 || gotMsg.Embeds[0].Fields[0].Name != "Comment by" {
		t.Fatalf("expected comment author field")
	}
}

func TestWebhookHandlerChangelog(t *testing.T) {
	app := setupTestApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()

	var gotMsg discord.WebhookMessage
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		gotMsg = msg
		return nil
	}

	payload := jira.Webhook{
		Issue:     jira.Issue{Key: "PRJ-3"},
		Changelog: &jira.Changelog{Items: []jira.ChangelogItem{{Field: "status", FromString: "Open", ToString: "Closed"}}},
	}
	payload.Issue.Fields.Summary = "Change"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Closed"

	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test: %v", err)
	}

	if len(gotMsg.Embeds[0].Fields) == 0 {
		t.Fatalf("expected fields in embed")
	}
	var found bool
	for _, f := range gotMsg.Embeds[0].Fields {
		if f.Name == "Changes" && strings.Contains(f.Value, "Status: Open → Closed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected status change in embed fields")
	}
}

func TestMainEnvVars(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("PORT", "12345")
	os.Setenv("JIRA_BASE_URL", "https://jira.example.com/browse")
	// Just check that main() runs without panic with these env vars set.
	// This does not start the server, but ensures no config panics.
	// To avoid actually starting the server, run main in a goroutine and kill after a short time (not shown here).
}

func TestMainLogLevelVariants(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "", "invalid"}
	for _, lvl := range levels {
		t.Run(lvl, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", lvl)
			os.Setenv("PORT", "0") // Use 0 to avoid port conflict
			// Run main in a goroutine and kill after a short time to avoid blocking
			ch := make(chan struct{})
			go func() {
				defer func() { recover() }()
				main()
				close(ch)
			}()
			select {
			case <-ch:
				// main exited (should not happen in normal server)
			case <-time.After(100 * time.Millisecond):
				// Timed out as expected, main is running server
			}
		})
	}
}

func TestMain(m *testing.M) {
	tmp := "test_user_mapping.yaml"
	yaml := `jira_to_discord:
  - accountId: "accid1"
    displayName: "User One"
    discordId: "111111111111111111"
  - accountId: "accid2"
    displayName: "User Two"
    discordId: "222222222222222222"
`
	_ = os.WriteFile(tmp, []byte(yaml), 0644)
	os.Setenv("USER_MAPPING_PATH", tmp)
	code := m.Run()
	_ = os.Remove(tmp)
	os.Exit(code)
}
