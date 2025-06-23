package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"jira-discord-webhook/internal/discord"
	"jira-discord-webhook/internal/jira"
)

func setupApp() *fiber.App {
	app := fiber.New()
	app.Post("/webhook", WebhookHandler)
	return app
}

func TestWebhookHandler_Success(t *testing.T) {
	app := setupApp()
	os.Setenv("JIRA_BASE_URL", "https://jira.example.com/browse")
	defer os.Unsetenv("JIRA_BASE_URL")

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
	payload.Issue.Fields.Summary = "Test issue"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Open"

	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_BadRequest(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWebhookHandler_DiscordError(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		return fiber.ErrInternalServerError
	}

	payload := jira.Webhook{
		Issue: jira.Issue{
			Key: "PRJ-1",
		},
	}
	payload.Issue.Fields.Summary = "Test issue"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Open"

	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWebhookHandler_MissingBaseURL(t *testing.T) {
	app := setupApp()
	os.Unsetenv("JIRA_BASE_URL")
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{
		Issue: jira.Issue{Key: "PRJ-1"},
	}
	payload.Issue.Fields.Summary = "Test issue"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Open"
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_CommentAndChangelog(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var gotMsg discord.WebhookMessage
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		gotMsg = msg
		return nil
	}
	payload := jira.Webhook{
		Issue:     jira.Issue{Key: "PRJ-4"},
		Comment:   &jira.Comment{},
		Changelog: &jira.Changelog{Items: []jira.ChangelogItem{{Field: "status", FromString: "Open", ToString: "Closed"}}},
	}
	payload.Issue.Fields.Summary = "Comment and Change issue"
	payload.Issue.Fields.Description = "desc"
	payload.Issue.Fields.Priority.Name = "High"
	payload.Issue.Fields.Assignee.DisplayName = "Bob"
	payload.Issue.Fields.Issuetype.Name = "Task"
	payload.Issue.Fields.Status.Name = "Closed"
	payload.Comment.Body = "needs work"
	payload.Comment.Author.DisplayName = "Alice"
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Contains(t, gotMsg.Embeds[0].Description, "needs work")
	var hasAuthor, hasChange bool
	for _, f := range gotMsg.Embeds[0].Fields {
		if f.Name == "Comment by" && f.Value == "Alice" {
			hasAuthor = true
		}
		if f.Name == "Changes" && f.Value == "Status: Open → Closed" {
			hasChange = true
		}
	}
	require.True(t, hasAuthor, "expected comment author field")
	require.True(t, hasChange, "expected changelog field")
}
