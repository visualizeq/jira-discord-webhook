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
		{"link", `[foo|http://bar]`, `[TABLE Content]`},
		{"bold", `+bold+`, `**bold**`},
		{"italic", `*italic*`, `_italic_`},
		{"underline", `_underline_`, `__underline__`},
		{"monospace", `{{code}}`, "`code`"},
		{"strikethrough", `-strike-`, `~~strike~~`},
		{"blockquote", "bq. quote", "> quote"},
		{"color", `{color:red}red text{color}`, `red text`},
		{"noformat", `{noformat}abc{noformat}`, "```\nabc\n```"},
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
		{"heading", "h2. Title", "## Title"},
		{"bullet", "* item", "- item"},
		{"numbered", "# item", "1. item"},
		{"hr", "----", "---"},
		{"mention", "[~bob]", "@bob"},
		{"attachment", "[^file.txt]", "file.txt"},
		{"image", "!pic.png!", "![](pic.png)"},
		{"urlInCodeBlock", "{code}http://example.com{code}", "```\nhttp://example.com\n```"},
		{"urlInInlineCode", "`http://example.com`", "`http://example.com`"},
		{"formattingInNoformat", "{noformat}*bold*{noformat}", "```\n*bold*\n```"},
		{"quote", `{quote}line1\nline2{quote}`, "> line1\n> line2"},
		{"tableHeader", "||A||B||", "[TABLE Content]"},
		{"tableRow", "|1|2|", "[TABLE Content]"},
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

func TestJiraToMarkdown_Mixed(t *testing.T) {
	in := strings.Join([]string{
		"h2. Heading",
		"----",
		"* +bold+ _underline_ -strike-",
		"# item",
		"bq. quote",
		"{color:red}colored{color}",
		"[link|http://example.com]",
		"{code:go}fmt.Println(1){code}",
		"{noformat}*literal*{noformat}",
		"{quote}line1",
		"line2{quote}",
		"{panel:title=Title}panel line{panel}",
		"!img.png!",
		"[^file.txt]",
		"[~bob]",
		"||A||B||",
		"|1|2|",
	}, "\n")

	want := strings.Join([]string{
		"## Heading",
		"---",
		"- **bold** __underline__ ~~strike~~",
		"1. item",
		"> quote",
		"colored",
		"[TABLE Content]",
		"```go",
		"fmt.Println(1)",
		"```",
		"```",
		"*literal*",
		"```",
		"> line1",
		"> line2",
		"> **Title**",
		"> panel line",
		"![](img.png)",
		"file.txt",
		"@bob",
		"[TABLE Content]",
	}, "\n")

	got := JiraToMarkdown(in)
	if got != want {
		t.Errorf("mixed format\ngot:\n%q\nwant:\n%q", got, want)
	}
}
