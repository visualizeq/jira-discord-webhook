package jira

import (
	"strings"
	"testing"
)

func TestJiraToMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"link", `[foo|http://bar]`, `[foo](http://bar)`},
		{"bold", `+bold+`, `**bold**`},
		{"italic", `*italic*`, `_italic_`},
		{"underline", `_underline_`, `__underline__`},
		{"monospace", `{{code}}`, "`code`"},
		{"strikethrough", `-strike-`, `~~strike~~`},
		{"blockquote", "bq. quote", "> quote"},
		{"color", `{color:red}red text{color}`, `red text`},
		{"noformat", `{noformat}abc{noformat}`, "```abc```"},
		{"panel", `{panel:title=Title}line1\nline2{panel}`,
			"> **Title**\n> line1\n> line2"},
		{"codeBlock", `{code:go}fmt.Println(1){code}`,
			"```go\nfmt.Println(1)\n```"},
		{"codeBlockNoLang", `{code}fmt.Println(1){code}`,
			"```\nfmt.Println(1)\n```"},
		{"superscript", `^sup^`, `^sup^`},
		{"subscript", `~sub~`, `~sub~`},
		{"strikethroughInCodeSpan", "`a-b-c-d-e.abc.com`", "`a-b-c-d-e.abc.com`"},
		{"strikethroughInCodeBlock", "```a-b-c-d-e.abc.com```", "```a-b-c-d-e.abc.com```"},
		{"strikethroughOutsideCode", "a-b-c-d-e.abc.com", "a-b-c-d-e.abc.com"},
		{"strikethroughDate", "2025-06-03", "2025-06-03"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := JiraToMarkdown(strings.ReplaceAll(tc.in, "\\n", "\n"))
			want := strings.ReplaceAll(tc.out, "\\n", "\n")
			if got != want {
				t.Errorf("input: %q\ngot:  %q\nwant: %q", tc.in, got, want)
			}
		})
	}
}
