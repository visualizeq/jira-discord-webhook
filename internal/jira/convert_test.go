package jira

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadWebhook(t *testing.T, name string) Webhook {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}
	var w Webhook
	if err := json.Unmarshal(b, &w); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return w
}

func TestToDiscordMessageIssue(t *testing.T) {
	w := loadWebhook(t, "issue.json")
	msg := ToDiscordMessage(w, "https://example.com/browse")
	if msg.Embeds[0].Title != "PRJ-1: Test issue" {
		t.Fatalf("unexpected title: %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].URL != "https://example.com/browse/PRJ-1" {
		t.Fatalf("unexpected url: %s", msg.Embeds[0].URL)
	}
}

func TestToDiscordMessageComment(t *testing.T) {
	w := loadWebhook(t, "comment.json")
	msg := ToDiscordMessage(w, "")
	if msg.Embeds[0].Description != "looks good" {
		t.Fatalf("expected comment body")
	}
	var found bool
	for _, f := range msg.Embeds[0].Fields {
		if f.Name == "Comment by" && f.Value == "Alice" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing comment author field")
	}
}

func TestToDiscordMessageChangelog(t *testing.T) {
	w := loadWebhook(t, "changelog.json")
	msg := ToDiscordMessage(w, "")
	var found bool
	for _, f := range msg.Embeds[0].Fields {
		if f.Name == "Changes" &&
			f.Value == "Status: Open → Closed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected status change field")
	}
}

func TestToDiscordMessageCommentChangelog(t *testing.T) {
	w := loadWebhook(t, "comment_changelog.json")
	msg := ToDiscordMessage(w, "")
	if msg.Embeds[0].Description != "needs work" {
		t.Fatalf("expected comment body")
	}
	var hasAuthor, hasChange bool
	for _, f := range msg.Embeds[0].Fields {
		if f.Name == "Comment by" && f.Value == "Alice" {
			hasAuthor = true
		}
		if f.Name == "Changes" && f.Value == "Status: Open → Closed" {
			hasChange = true
		}
	}
	if !hasAuthor || !hasChange {
		t.Fatalf("expected comment author and change fields")
	}
}
