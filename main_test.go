package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestWebhookHandlerBadJSON(t *testing.T) {
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
