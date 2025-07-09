package utils

import "testing"

func TestProtectDomainsAndFiles(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"bulletDomain", "* a-b-c-d-e.abc.com (111.222.232.98)", "* `a-b-c-d-e.abc.com` (`111.222.232.98`)"},
		{"fullLineDomain", "a-b-c-d-e.abc.com", "`a-b-c-d-e.abc.com`"},
		{"subdomainTLD", "foo.bar.co.uk", "`foo.bar.co.uk`"},
		{"filenameMultiDot", "archive.tar.gz", "`archive.tar.gz`"},
		{"inlineDomain", "Visit a-b-c-d-e.abc.com for info", "Visit `a-b-c-d-e.abc.com` for info"},
		{"multipleDomains", "a-b-c-d-e.abc.com and x.y.z.com", "`a-b-c-d-e.abc.com` and `x.y.z.com`"},
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
		{"jiraImageAlreadyWrapped", "`Screenshot 2025-04-17 at 14.39.17.png`", "`Screenshot 2025-04-17 at 14.39.17.png`"},
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
