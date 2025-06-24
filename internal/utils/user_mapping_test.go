package utils

import (
	"os"
	"testing"
)

func TestLoadUserMappingAndDiscordMention(t *testing.T) {
	// Prepare a temporary YAML file
	yamlContent := `jira_to_discord:
  - accountId: "accid1"
    displayName: "User One"
    discordId: "111111111111111111"
  - accountId: "accid2"
    displayName: "User Two"
    discordId: "222222222222222222"
`
	tmpFile, err := os.CreateTemp("", "user_mapping_test_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}
	tmpFile.Close()

	// Load mapping
	if err := LoadUserMapping(tmpFile.Name()); err != nil {
		t.Fatalf("LoadUserMapping failed: %v", err)
	}

	// Test by accountId
	if got := DiscordMentionForJiraUser("accid1"); got != "<@111111111111111111>" {
		t.Errorf("expected <@111111111111111111>, got %s", got)
	}
	// Test by displayName
	if got := DiscordMentionForJiraUser("User Two"); got != "<@222222222222222222>" {
		t.Errorf("expected <@222222222222222222>, got %s", got)
	}
	// Test fallback
	if got := DiscordMentionForJiraUser("unknown"); got != "unknown" {
		t.Errorf("expected unknown, got %s", got)
	}
}

func TestReplaceJiraMentionsWithDiscord(t *testing.T) {
	// Setup a fake mapping
	jiraToDiscord = UserMapping{
		JiraToDiscord: []JiraUserMapping{
			{AccountID: "accid1", DisplayName: "User One", DiscordID: "111111111111111111"},
			{AccountID: "accid2", DisplayName: "User Two", DiscordID: "222222222222222222"},
		},
	}

	// Test single accountId mention
	in := "Hello [~accountid:accid1]!"
	want := "Hello <@111111111111111111>!"
	if got := ReplaceJiraMentionsWithDiscord(in); got != want {
		t.Errorf("singleAccountIdMention: got %q, want %q", got, want)
	}

	// Test multiple accountId mentions
	in = "[~accountid:accid1] and [~accountid:accid2] are here."
	want = "<@111111111111111111> and <@222222222222222222> are here."
	if got := ReplaceJiraMentionsWithDiscord(in); got != want {
		t.Errorf("multipleAccountIdMentions: got %q, want %q", got, want)
	}

	// Test unknown accountId
	in = "Hi [~accountid:unknown]!"
	want = "Hi unknown!" // The current implementation falls back to the key if not found
	if got := ReplaceJiraMentionsWithDiscord(in); got != want {
		t.Errorf("unknownAccountId: got %q, want %q", got, want)
	}
}

func TestProtectDomains(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"bulletDomain", "* a-b-c-d-e.abc.com (111.222.232.98)", "* `a-b-c-d-e.abc.com` (`111.222.232.98`)"},
		{"fullLineDomain", "a-b-c-d-e.abc.com", "```a-b-c-d-e.abc.com```"},
		{"inlineDomain", "Visit a-b-c-d-e.abc.com for info", "Visit `a-b-c-d-e.abc.com` for info"},
		{"multipleDomains", "a-b-c-d-e.abc.com and x.y.z.com", "`a-b-c-d-e.abc.com` and `x.y.z.com`"},
		{"noDomain", "hello world", "hello world"},
		{"domainInBulletWithExtraText", "* see a-b-c-d-e.abc.com for info", "* see `a-b-c-d-e.abc.com` for info"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ProtectDomains(tc.in)
			if got != tc.out {
				t.Errorf("input: %q\ngot:  %q\nwant: %q", tc.in, got, tc.out)
			}
		})
	}
}
