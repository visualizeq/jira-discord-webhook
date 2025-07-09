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

func TestLoadUserMapping_ErrorCases(t *testing.T) {
	// Test file not found
	err := LoadUserMapping("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}

	// Test invalid YAML
	tmpFile, err := os.CreateTemp("", "user_mapping_test_invalid_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write([]byte("not: valid: yaml: : :")); err != nil {
		t.Fatalf("failed to write temp yaml: %v", err)
	}
	tmpFile.Close()
	if err := LoadUserMapping(tmpFile.Name()); err == nil {
		t.Error("expected error for invalid yaml")
	}
}

func TestDiscordMentionForJiraUser_EmptyMapping(t *testing.T) {
	jiraToDiscord = UserMapping{}
	if got := DiscordMentionForJiraUser("anyone"); got != "anyone" {
		t.Errorf("expected fallback to key, got %q", got)
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

	// Test simple user mention (fallback to @user)
	in = "Hello [~bob]!"
	want = "Hello @bob!"
	if got := ReplaceJiraMentionsWithDiscord(in); got != want {
		t.Errorf("simpleUserMention: got %q, want %q", got, want)
	}

	// Test simple user mention with mapping
	jiraToDiscord.JiraToDiscord = append(jiraToDiscord.JiraToDiscord, JiraUserMapping{
		AccountID: "bob", DisplayName: "Bob", DiscordID: "333333333333333333",
	})
	in = "Hello [~bob]!"
	want = "Hello <@333333333333333333>!"
	if got := ReplaceJiraMentionsWithDiscord(in); got != want {
		t.Errorf("simpleUserMentionWithMapping: got %q, want %q", got, want)
	}
}

func TestReplaceJiraMentionsWithDiscord_NoMentions(t *testing.T) {
	jiraToDiscord = UserMapping{}
	in := "No mentions here."
	if got := ReplaceJiraMentionsWithDiscord(in); got != in {
		t.Errorf("expected unchanged, got %q", got)
	}
}

// Moved to helper_test.go
/*
func TestProtectDomainsAndFiles(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"bulletDomain", "* a-b-c-d-e.abc.com (111.222.232.98)", "* `a-b-c-d-e.abc.com` (`111.222`.`232.98`)"},
		{"fullLineDomain", "a-b-c-d-e.abc.com", "`a-b-c-d-e.abc.com`"},
		{"subdomainTLD", "foo.bar.co.uk", "`foo.bar`.`co.uk`"},
		{"filenameMultiDot", "archive.tar.gz", "`archive.tar.gz`"},
		{"inlineDomain", "Visit a-b-c-d-e.abc.com for info", "Visit `a-b-c-d-e.abc.com` for info"},
		{"multipleDomains", "a-b-c-d-e.abc.com and x.y.z.com", "`a-b-c-d-e.abc.com` and `x.y`.`z.com`"},
		{"noDomain", "hello world", "hello world"},
		{"domainInBulletWithExtraText", "* see a-b-c-d-e.abc.com for info", "* see `a-b-c-d-e.abc.com` for info"},
		{"filename", "This is file.txt", "This is `file.txt`"},
		{"filenameAndDomain", "file.txt and a-b-c-d-e.abc.com", "`file.txt` and `a-b-c-d-e.abc.com`"},
		{"filenameInCode", "`file.txt`", "`file.txt`"},
		{"urlShouldNotWrap", "See https://jira.example.com/browse/PRJ-3", "See https://jira.example.com/browse/PRJ-3"},
		{"bareDomainShouldWrap", "See jira.example.com for info", "See `jira.example.com` for info"},
		{"urlAndBareDomain", "See https://jira.example.com and jira.example.com", "See https://jira.example.com and `jira.example.com`"},
		{"filenameDoubleUnderscore", "run move__bank__cus_mapping.sh", "run `move_bank_cus_mapping.sh`"},
		{"filenameWithPunct", "edit file.txt, then exit", "edit `file.txt`, then exit"},
		{"urlWithPunct", "Go to https://jira.example.com/.", "Go to https://jira.example.com/."},
		{"domainAtStart", "abc.com is up", "`abc.com` is up"},
		{"domainAtEnd", "see abc.com", "see `abc.com`"},
		{"domainWithPunct", "abc.com, xyz.com.", "`abc.com`, `xyz.com`."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ProtectDomainsAndFiles(tc.in)
			if got != tc.out {
				t.Errorf("input: %q\ngot:  %q\nwant: %q", tc.in, got, tc.out)
			}
		})
	}
}
*/

func TestIsInRanges(t *testing.T) {
	ranges := [][2]int{{5, 10}, {15, 20}}
	cases := []struct {
		start, end int
		want       bool
	}{
		{5, 7, true},    // inside first range
		{8, 10, true},   // end at range end
		{10, 12, false}, // outside
		{15, 18, true},  // inside second range
		{0, 4, false},   // before all
	}
	for _, c := range cases {
		got := isInRanges(c.start, c.end, ranges)
		if got != c.want {
			t.Errorf("isInRanges(%d, %d, %v) = %v, want %v", c.start, c.end, ranges, got, c.want)
		}
	}
}

func TestIsInBackticks(t *testing.T) {
	cases := []struct {
		line, domain string
		want         bool
	}{
		{"`foo-bar.com`", "foo-bar.com", true},
		{"foo-bar.com", "foo-bar.com", false},
		{"prefix `foo-bar.com` suffix", "foo-bar.com", true},
		{"prefix foo-bar.com suffix", "foo-bar.com", false},
	}
	for _, c := range cases {
		got := isInBackticks(c.line, c.domain)
		if got != c.want {
			t.Errorf("isInBackticks(%q, %q) = %v, want %v", c.line, c.domain, got, c.want)
		}
	}
}
