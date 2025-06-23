package utils

import (
	"os"
	"testing"
)

func TestLoadUserMappingAndMention(t *testing.T) {
	// Prepare a temp TOML file
	toml := `[jira_to_discord]
"Alice" = "123"
"Bob" = "456"
`
	tmp := "test_user_mapping.toml"
	if err := os.WriteFile(tmp, []byte(toml), 0644); err != nil {
		t.Fatalf("failed to write temp toml: %v", err)
	}
	defer os.Remove(tmp)

	if err := LoadUserMapping(tmp); err != nil {
		t.Fatalf("LoadUserMapping failed: %v", err)
	}

	tests := []struct {
		jiraName string
		expect   string
	}{
		{"Alice", "<@123>"},
		{"Bob", "<@456>"},
		{"Unknown", "Unknown"},
	}
	for _, tc := range tests {
		got := DiscordMentionForJiraUser(tc.jiraName)
		if got != tc.expect {
			t.Errorf("mention for %q: got %q, want %q", tc.jiraName, got, tc.expect)
		}
	}
}
