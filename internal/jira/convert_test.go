package jira

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	os.Unsetenv("ISSUE_COLOR")
	w := loadWebhook(t, "issue.json")
	msg := ToDiscordMessage(w, "https://example.com/browse")
	if msg.Embeds[0].Title != "PRJ-1: Test issue" {
		t.Fatalf("unexpected title: %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].URL != "https://example.com/browse/PRJ-1" {
		t.Fatalf("unexpected url: %s", msg.Embeds[0].URL)
	}
	if msg.Embeds[0].Color != issueColor {
		t.Fatalf("unexpected color: %d", msg.Embeds[0].Color)
	}
}

func TestToDiscordMessageComment(t *testing.T) {
	os.Unsetenv("COMMENT_COLOR")
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
	if msg.Embeds[0].Color != commentColor {
		t.Fatalf("unexpected color: %d", msg.Embeds[0].Color)
	}
}

func TestToDiscordMessageChangelog(t *testing.T) {
	os.Unsetenv("CHANGELOG_COLOR")
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
	if msg.Embeds[0].Color != changelogColor {
		t.Fatalf("unexpected color: %d", msg.Embeds[0].Color)
	}
}

func TestToDiscordMessageCommentChangelog(t *testing.T) {
	os.Unsetenv("COMMENT_CHANGELOG_COLOR")
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
	if msg.Embeds[0].Color != commentChangelogColor {
		t.Fatalf("unexpected color: %d", msg.Embeds[0].Color)
	}
}

func TestToDiscordMessageColorFromEnv(t *testing.T) {
	os.Setenv("ISSUE_COLOR", "0x123456")
	defer os.Unsetenv("ISSUE_COLOR")
	w := loadWebhook(t, "issue.json")
	msg := ToDiscordMessage(w, "")
	if msg.Embeds[0].Color != 0x123456 {
		t.Fatalf("env color not applied")
	}
}

func TestToDiscordMessage_EmptyFields(t *testing.T) {
	w := Webhook{
		Issue: Issue{Key: "PRJ-EMPTY"},
	}
	msg := ToDiscordMessage(w, "")
	if msg.Embeds[0].Title != "PRJ-EMPTY: " {
		t.Fatalf("unexpected title: %s", msg.Embeds[0].Title)
	}
	if msg.Embeds[0].Description != "" {
		t.Fatalf("expected empty description")
	}
}

func TestToDiscordMessage_LongFields(t *testing.T) {
	long := strings.Repeat("A", 300)
	w := Webhook{
		Issue: Issue{Key: long},
	}
	w.Issue.Fields.Summary = long
	w.Issue.Fields.Description = long
	msg := ToDiscordMessage(w, "")
	if len(msg.Embeds[0].Title) > 256 {
		t.Fatalf("title too long")
	}
	if len(msg.Embeds[0].Description) > 4096 {
		t.Fatalf("description too long")
	}
}
