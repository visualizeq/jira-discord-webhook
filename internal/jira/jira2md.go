// Force rebuild: 2025-06-23
package jira

import (
	"regexp"
	"strings"
)

// JiraToMarkdown converts Jira wiki markup to Markdown/Discord formatting.
// Example: [text|http://example.com] => [text](http://example.com)
func JiraToMarkdown(s string) string {
	// Convert {code} and {noformat} blocks to fenced code blocks first so
	// their contents are not processed by other transformations.
	codeBlockRE := regexp.MustCompile(`(?s)\{code(?::([a-zA-Z0-9_+-]+))?\}(.*?)\{code\}`)
	s = codeBlockRE.ReplaceAllStringFunc(s, func(m string) string {
		parts := codeBlockRE.FindStringSubmatch(m)
		lang := parts[1]
		content := strings.TrimSpace(parts[2])
		if lang != "" {
			return "```" + lang + "\n" + content + "\n```"
		}
		return "```\n" + content + "\n```"
	})
	noformatRE := regexp.MustCompile(`(?s)\{noformat\}(.*?)\{noformat\}`)
	s = noformatRE.ReplaceAllStringFunc(s, func(m string) string {
		parts := noformatRE.FindStringSubmatch(m)
		content := strings.TrimSpace(parts[1])
		return "```\n" + content + "\n```"
	})
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
		// Only match -text- that is surrounded by word boundaries and not part of a hostname or URL
		seg.text = regexp.MustCompile(`\B-([a-zA-Z0-9][^\s-]*[a-zA-Z0-9])-\B`).ReplaceAllString(seg.text, `~~$1~~`)
		// Headings: h1. text -> # text
		seg.text = regexp.MustCompile(`(?m)^h([1-6])\.\s+(.+)$`).ReplaceAllStringFunc(seg.text, func(m string) string {
			headerRE := regexp.MustCompile(`(?m)^h([1-6])\.\s+(.+)$`)
			parts := headerRE.FindStringSubmatch(m)
			if len(parts) == 3 {
				level := parts[1]
				n := 0
				switch level {
				case "1":
					n = 1
				case "2":
					n = 2
				case "3":
					n = 3
				case "4":
					n = 4
				case "5":
					n = 5
				case "6":
					n = 6
				}
				return strings.Repeat("#", n) + " " + parts[2]
			}
			return m
		})
		// Horizontal rule ---- -> ---
		seg.text = regexp.MustCompile(`(?m)^----+\s*$`).ReplaceAllString(seg.text, "---")
		// Bullet list: * item -> - item
		seg.text = regexp.MustCompile(`(?m)^[ \t]*\*\s+`).ReplaceAllString(seg.text, "- ")
		// Numbered list: # item -> 1. item
		seg.text = regexp.MustCompile(`(?m)^[ \t]*#\s+`).ReplaceAllString(seg.text, "1. ")
		// Blockquote: bq. text -> > text
		seg.text = regexp.MustCompile(`(?m)^bq\.\s+`).ReplaceAllString(seg.text, "> ")
		// Remove color markup: {color:red}text{color} -> text
		seg.text = regexp.MustCompile(`\{color:[^}]+\}(.*?)\{color\}`).ReplaceAllString(seg.text, `$1`)
		// {quote}...{quote} -> blockquote
		seg.text = regexp.MustCompile(`(?s)\{quote\}(.*?)\{quote\}`).ReplaceAllStringFunc(seg.text, func(m string) string {
			quoteRE := regexp.MustCompile(`(?s)\{quote\}(.*?)\{quote\}`)
			parts := quoteRE.FindStringSubmatch(m)
			if len(parts) == 2 {
				lines := strings.Split(strings.TrimSpace(parts[1]), "\n")
				for i, l := range lines {
					lines[i] = "> " + l
				}
				return strings.Join(lines, "\n")
			}
			return m
		})
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
		// Mentions: [~user] -> @user
		seg.text = regexp.MustCompile(`\[~([^\]]+)\]`).ReplaceAllString(seg.text, "@$1")
		// Attachment: [^file.ext] -> file.ext
		seg.text = regexp.MustCompile(`\[\^([^\]]+)\]`).ReplaceAllString(seg.text, "$1")
		// Image: !img.png! -> ![](img.png)
		seg.text = regexp.MustCompile(`!([^!]+)!`).ReplaceAllString(seg.text, "![]($1)")
		// Table header ||a||b|| -> | a | b |
		seg.text = regexp.MustCompile(`(?m)^\|\|(.+?)\|\|$`).ReplaceAllStringFunc(seg.text, func(m string) string {
			trimmed := strings.Trim(m, "|")
			cells := strings.Split(trimmed, "||")
			for i, c := range cells {
				cells[i] = strings.TrimSpace(c)
			}
			return "| " + strings.Join(cells, " | ") + " |"
		})
		// Table row |a|b| -> | a | b |
		seg.text = regexp.MustCompile(`(?m)^\|([^|].*?)\|$`).ReplaceAllStringFunc(seg.text, func(m string) string {
			trimmed := strings.Trim(m, "|")
			cells := strings.Split(trimmed, "|")
			for i, c := range cells {
				cells[i] = strings.TrimSpace(c)
			}
			return "| " + strings.Join(cells, " | ") + " |"
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
