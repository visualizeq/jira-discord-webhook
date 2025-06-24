// Force rebuild: 2025-06-23
package jira

import (
	"regexp"
	"strings"
)

// JiraToMarkdown converts Jira wiki markup to Markdown/Discord formatting.
// Example: [text|http://example.com] => [text](http://example.com)
func JiraToMarkdown(s string) string {
	// Helper: split into code and non-code segments
	type segment struct {
		text   string
		isCode bool
	}
	segments := make([]segment, 0)
	var buf strings.Builder
	inCode := false
	codeDelim := ""
	for i := 0; i < len(s); {
		if !inCode && strings.HasPrefix(s[i:], "```") {
			if buf.Len() > 0 {
				segments = append(segments, segment{buf.String(), false})
				buf.Reset()
			}
			inCode = true
			codeDelim = "```"
			buf.WriteString("```")
			i += 3
			continue
		}
		if !inCode && strings.HasPrefix(s[i:], "`") {
			if buf.Len() > 0 {
				segments = append(segments, segment{buf.String(), false})
				buf.Reset()
			}
			inCode = true
			codeDelim = "`"
			buf.WriteByte('`')
			i++
			continue
		}
		if inCode && strings.HasPrefix(s[i:], codeDelim) {
			buf.WriteString(codeDelim)
			i += len(codeDelim)
			segments = append(segments, segment{buf.String(), true})
			buf.Reset()
			inCode = false
			codeDelim = ""
			continue
		}
		buf.WriteByte(s[i])
		i++
	}
	if buf.Len() > 0 {
		segments = append(segments, segment{buf.String(), inCode})
	}

	for i, seg := range segments {
		if seg.isCode {
			continue
		}
		// Links: [text|url] -> [text](url)
		jiraLinkRE := regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
		seg.text = jiraLinkRE.ReplaceAllStringFunc(seg.text, func(m string) string {
			parts := jiraLinkRE.FindStringSubmatch(m)
			if len(parts) == 3 {
				return "[" + parts[1] + "](" + parts[2] + ")"
			}
			return m
		})
		// Underline: _text_ -> __text__ (run first)
		seg.text = regexp.MustCompile(`_([^_\n]+)_`).ReplaceAllStringFunc(seg.text, func(m string) string {
			if len(m) > 2 && m[0] == '_' && m[len(m)-1] == '_' {
				return "__" + m[1:len(m)-1] + "__"
			}
			return m
		})
		// Italic: *text* -> _text_ (run second)
		seg.text = regexp.MustCompile(`\*([^\*\n]+)\*`).ReplaceAllStringFunc(seg.text, func(m string) string {
			if len(m) > 2 && m[0] == '*' && m[len(m)-1] == '*' {
				return "_" + m[1:len(m)-1] + "_"
			}
			return m
		})
		// Bold: +text+ -> **text** (run last)
		seg.text = regexp.MustCompile(`\+([^\+\n]+)\+`).ReplaceAllStringFunc(seg.text, func(m string) string {
			if len(m) > 2 {
				return "**" + m[1:len(m)-1] + "**"
			}
			return m
		})
		// Monospace: {{text}} -> `text`
		seg.text = regexp.MustCompile(`\{\{(.*?)\}\}`).ReplaceAllString(seg.text, "`$1`")
		// Strikethrough: -text- -> ~~text~~ (only outside code)
		seg.text = regexp.MustCompile(`-(.*?)-`).ReplaceAllString(seg.text, `~~$1~~`)
		// Blockquote: bq. text -> > text
		seg.text = regexp.MustCompile(`(?m)^bq\.\s+`).ReplaceAllString(seg.text, "> ")
		// Remove color markup: {color:red}text{color} -> text
		seg.text = regexp.MustCompile(`\{color:[^}]+\}(.*?)\{color\}`).ReplaceAllString(seg.text, `$1`)
		// {noformat}...{noformat} -> ```...```
		seg.text = regexp.MustCompile(`(?s)\{noformat\}(.*?)\{noformat\}`).ReplaceAllString(seg.text, "```$1```")
		// {panel:title=Title}...{panel} -> > **Title**\n> ...
		seg.text = regexp.MustCompile(`(?s)\{panel:title=([^}]*)\}(.*?)\{panel\}`).ReplaceAllStringFunc(seg.text, func(m string) string {
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
		seg.text = regexp.MustCompile(`(?s)\{code(?::([a-zA-Z0-9_+-]+))?\}(.*?)\{code\}`).ReplaceAllStringFunc(seg.text, func(m string) string {
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
		segments[i] = seg
	}
	// Reassemble
	var out strings.Builder
	for _, seg := range segments {
		out.WriteString(seg.text)
	}
	return out.String()
}
