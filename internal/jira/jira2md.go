// Force rebuild: 2025-06-23
package jira

import (
	"regexp"
	"strings"
)

// JiraToMarkdown converts Jira wiki markup to Markdown/Discord formatting.
// Example: [text|http://example.com] => [text](http://example.com)
func JiraToMarkdown(s string) string {
	// Links: [text|url] -> [text](url)
	jiraLinkRE := regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
	s = jiraLinkRE.ReplaceAllStringFunc(s, func(m string) string {
		parts := jiraLinkRE.FindStringSubmatch(m)
		if len(parts) == 3 {
			return "[" + parts[1] + "](" + parts[2] + ")"
		}
		return m
	})
	// Underline: _text_ -> __text__ (run first)
	s = regexp.MustCompile(`_([^_\n]+)_`).ReplaceAllStringFunc(s, func(m string) string {
		if len(m) > 2 && m[0] == '_' && m[len(m)-1] == '_' {
			return "__" + m[1:len(m)-1] + "__"
		}
		return m
	})
	// Italic: *text* -> _text_ (run second)
	s = regexp.MustCompile(`\*([^\*\n]+)\*`).ReplaceAllStringFunc(s, func(m string) string {
		if len(m) > 2 && m[0] == '*' && m[len(m)-1] == '*' {
			return "_" + m[1:len(m)-1] + "_"
		}
		return m
	})
	// Bold: +text+ -> **text** (run last)
	s = regexp.MustCompile(`\+([^\+\n]+)\+`).ReplaceAllStringFunc(s, func(m string) string {
		if len(m) > 2 {
			return "**" + m[1:len(m)-1] + "**"
		}
		return m
	})
	// Monospace: {{text}} -> `text`
	s = regexp.MustCompile(`\{\{(.*?)\}\}`).ReplaceAllString(s, "`$1`")
	// Strikethrough: -text- -> ~~text~~
	s = regexp.MustCompile(`-(.*?)-`).ReplaceAllString(s, `~~$1~~`)
	// Blockquote: bq. text -> > text
	s = regexp.MustCompile(`(?m)^bq\.\s+`).ReplaceAllString(s, "> ")
	// Remove color markup: {color:red}text{color} -> text
	s = regexp.MustCompile(`\{color:[^}]+\}(.*?)\{color\}`).ReplaceAllString(s, `$1`)
	// {noformat}...{noformat} -> ```...```
	s = regexp.MustCompile(`(?s)\{noformat\}(.*?)\{noformat\}`).ReplaceAllString(s, "```$1```")
	// {panel:title=Title}...{panel} -> > **Title**\n> ...
	s = regexp.MustCompile(`(?s)\{panel:title=([^}]*)\}(.*?)\{panel\}`).ReplaceAllStringFunc(s, func(m string) string {
		panelRE := regexp.MustCompile(`(?s)\{panel:title=([^}]*)\}(.*?)\{panel\}`)
		parts := panelRE.FindStringSubmatch(m)
		if len(parts) == 3 {
			lines := strings.Split(strings.TrimSpace(parts[2]), "\n")
			for i, l := range lines {
				lines[i] = "> " + l
			}
			return "> **" + parts[1] + "**\n" + strings.Join(lines, "\n")
		}
		return m
	})
	// {code[:lang]}...{code} -> ```lang\n...```
	s = regexp.MustCompile(`(?s)\{code(?::([a-zA-Z0-9_+-]+))?\}(.*?)\{code\}`).ReplaceAllStringFunc(s, func(m string) string {
		codeRE := regexp.MustCompile(`(?s)\{code(?::([a-zA-Z0-9_+-]+))?\}(.*?)\{code\}`)
		parts := codeRE.FindStringSubmatch(m)
		if len(parts) == 3 {
			lang := parts[1]
			if lang != "" {
				return "```" + lang + "\n" + strings.TrimSpace(parts[2]) + "\n```"
			}
			return "```\n" + strings.TrimSpace(parts[2]) + "\n```"
		}
		return m
	})
	// Superscript: ^text^ (no Discord equivalent, keep as is)
	// Subscript: ~text~ (no Discord equivalent, keep as is)
	return s
}
