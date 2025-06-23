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

func TestWebhookHandler_MinimalPayload(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: "PRJ-2"}}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_ExtraFields(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := map[string]interface{}{
		"issue": map[string]interface{}{"key": "PRJ-3"},
		"extra": "field",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_NilCommentChangelog(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{
		Issue:     jira.Issue{Key: "PRJ-5"},
		Comment:   nil,
		Changelog: nil,
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_DiscordCustomError(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		return fiber.ErrBadGateway
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: "PRJ-6"}}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWebhookHandler_LargeValidPayload(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: "PRJ-LARGE"}}
	payload.Issue.Fields.Summary = string(bytes.Repeat([]byte{'a'}, 10000))
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_EmptyIssueKey(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: ""}}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_EmptyFields(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: "PRJ-EMPTY"}}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_EmptyCommentBody(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{
		Issue:   jira.Issue{Key: "PRJ-COMMENT"},
		Comment: &jira.Comment{}, // No body, no author
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_EmptyChangelogItems(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{
		Issue:     jira.Issue{Key: "PRJ-CHANGELOG"},
		Changelog: &jira.Changelog{Items: []jira.ChangelogItem{}},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_CommentNoAuthor(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	var called bool
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		called = true
		return nil
	}
	payload := jira.Webhook{
		Issue:   jira.Issue{Key: "PRJ-NOAUTHOR"},
		Comment: &jira.Comment{Body: "Some comment"}, // No author
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.True(t, called, "SendFunc should be called")
}

func TestWebhookHandler_BadRequestPayload(t *testing.T) {
	app := setupApp()
	// Invalid JSON
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWebhookHandler_DiscordSendError(t *testing.T) {
	app := setupApp()
	original := discord.SendFunc
	defer func() { discord.SendFunc = original }()
	discord.SendFunc = func(msg discord.WebhookMessage) error {
		return fiber.ErrInternalServerError
	}
	payload := jira.Webhook{Issue: jira.Issue{Key: "PRJ-ERR"}}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWebhookHandler_EmptyBody(t *testing.T) {
	app := setupApp()
	req := httptest.NewRequest("POST", "/webhook", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWebhookHandler_UnsupportedMethod(t *testing.T) {
	app := setupApp()
	// Fiber returns 405 Method Not Allowed for unsupported methods on a registered route
	resp, err := app.Test(httptest.NewRequest("GET", "/webhook", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
}

func TestWebhookHandler_UnregisteredRoute(t *testing.T) {
	app := setupApp()
	resp, err := app.Test(httptest.NewRequest("POST", "/notfound", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
