// Force rebuild: 2025-06-23
package jira

import (
	"regexp"
	"strings"
)

// JiraToMarkdown converts Jira wiki markup to Markdown/Discord formatting.
// Example: [text|http://example.com] => [text](http://example.com)
func JiraToMarkdown(s string) string {
	// Remove all table lines (headers and rows)
	tableHeaderRE := regexp.MustCompile(`(?m)^\|\|.*\|\|\s*$`)
	tableRowRE := regexp.MustCompile(`(?m)^\|[^|].*\|\s*$`)

	// Remove Jira links inside table lines (including markdown links)
	jiraLinkInTableRE := regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
	markdownLinkInTableRE := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	tableLineRE := regexp.MustCompile(`(?m)^(\|\|.*\|\|\s*$|\|[^|].*\|\s*$)`)
	s = tableLineRE.ReplaceAllStringFunc(s, func(line string) string {
		line = jiraLinkInTableRE.ReplaceAllString(line, "$1")
		line = markdownLinkInTableRE.ReplaceAllString(line, "$1")
		return line
	})

	s = tableHeaderRE.ReplaceAllString(s, "")
	s = tableRowRE.ReplaceAllString(s, "")
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
		// Links: [text|url] -> [text](url), but remove protocol from text if text is a URL
		jiraLinkRE := regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
		seg.text = jiraLinkRE.ReplaceAllStringFunc(seg.text, func(m string) string {
			parts := jiraLinkRE.FindStringSubmatch(m)
			if len(parts) == 3 {
				text := parts[1]
				url := parts[2]
				// If text is a URL, strip protocol
				if strings.HasPrefix(text, "http://") {
					text = text[len("http://"):]
				} else if strings.HasPrefix(text, "https://") {
					text = text[len("https://"):]
				}
				return "[" + text + "](" + url + ")"
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
		// Removed table header and row reconstruction
		// Attachment: [^file.ext] -> file.txt
		seg.text = regexp.MustCompile(`\[\^([^\]]+)\]`).ReplaceAllString(seg.text, "$1")
		// Image: !img.png! -> ![](img.png)
		seg.text = regexp.MustCompile(`!([^!]+)!`).ReplaceAllString(seg.text, "![]($1)")
		// Links: [text|url] -> [text](url)
		jiraLinkRE = regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
		seg.text = jiraLinkRE.ReplaceAllStringFunc(seg.text, func(m string) string {
			parts := jiraLinkRE.FindStringSubmatch(m)
			if len(parts) == 3 {
				return "[" + parts[1] + "](" + parts[2] + ")"
			}
			return m
		})
		// Mentions: [~user] -> @user
		seg.text = regexp.MustCompile(`\[~([^\]]+)\]`).ReplaceAllString(seg.text, "@$1")
		// Now wrap bare domains/links in inline code, but skip inside []() and ![]() and also skip inside link text
		// Split on Markdown links and Jira links, only wrap in non-link segments
		// Allow for optional leading whitespace and list markers before the link/image/Jira link
		linkOrImageOrJiraLinkRE := regexp.MustCompile(`[ \t\-*]*!?\[[^\]]*\]\([^\)]*\)|[ \t\-*]*\[[^\]|]+\|[^\]]+\]`)
		parts := linkOrImageOrJiraLinkRE.Split(seg.text, -1)
		matches := linkOrImageOrJiraLinkRE.FindAllStringIndex(seg.text, -1)
		var rebuilt strings.Builder
		for i, part := range parts {
			// Refined: Only wrap domains that are surrounded by whitespace, start/end, or punctuation
			urlRE := regexp.MustCompile(`(?m)(^|[\s>\(\[\{])((?:https?|ftp)://[\w\-._~:/?#\[\]@!$&'()*+,;=%]+)([\s\.,;:!?\)\]\}\n]|$)`)
			part = urlRE.ReplaceAllStringFunc(part, func(m string) string {
				matches := urlRE.FindStringSubmatch(m)
				if len(matches) != 4 {
					return m
				}
				prefix := matches[1]
				url := matches[2]
				suffix := matches[3]
				if strings.HasPrefix(url, "`") && strings.HasSuffix(url, "`") {
					return m // already wrapped
				}
				return prefix + "`" + url + "`" + suffix
			})
			domainRE := regexp.MustCompile(`(?m)(^|[\s>\(\[\{])([a-zA-Z0-9][a-zA-Z0-9\-]*\.[a-zA-Z]{2,})([\s\.,;:!?\)\]\}\n]|$)`)
			part = domainRE.ReplaceAllStringFunc(part, func(m string) string {
				matches := domainRE.FindStringSubmatch(m)
				if len(matches) != 4 {
					return m
				}
				prefix := matches[1]
				domain := matches[2]
				suffix := matches[3]
				if strings.HasPrefix(domain, "`") && strings.HasSuffix(domain, "`") {
					return m // already wrapped
				}
				if strings.Contains(domain, "@") {
					return m // email
				}
				// Only wrap if not part of a larger word (e.g., not a-b-c-d-e.abc.com)
				if len(prefix) > 0 && (prefix[len(prefix)-1] == '-' || prefix[len(prefix)-1] == '.') {
					return m
				}
				if len(suffix) > 0 && (suffix[0] == '-' || suffix[0] == '.') {
					return m
				}
				// Do not wrap if it looks like a filename (e.g., file.txt, file-1.txt)
				filenameRE := regexp.MustCompile(`^[\w\-.]+\.[a-zA-Z0-9]+$`)
				if filenameRE.MatchString(domain) {
					return m
				}
				return prefix + "`" + domain + "`" + suffix
			})
			rebuilt.WriteString(part)
			if i < len(matches) {
				// Add the matched link/image/Jira link back untouched
				rebuilt.WriteString(seg.text[matches[i][0]:matches[i][1]])
			}
		}
		seg.text = rebuilt.String()
		segments[i] = seg
	}
	// Reassemble
	var out strings.Builder
	for _, seg := range segments {
		out.WriteString(seg.text)
	}
	s = out.String()
	// Replace any block of consecutive table-like lines with a single [TABLE Content]
	tableLikeLineRE := regexp.MustCompile(`^[ \t]*\|.*$|^.*\|[ \t]*$|^[ \t]*\[.*\]\(.*\)[ \t]*\|?[ \t]*$`)
	lines := strings.Split(s, "\n")
	var result []string
	inTable := false
	for i := 0; i < len(lines); {
		if tableLikeLineRE.MatchString(lines[i]) {
			// Start of a table block
			if !inTable {
				result = append(result, "[TABLE Content]")
				inTable = true
			}
			// Skip all consecutive table-like lines
			for i < len(lines) && tableLikeLineRE.MatchString(lines[i]) {
				i++
			}
			continue
		}
		inTable = false
		result = append(result, lines[i])
		i++
	}
	s = strings.Join(result, "\n")
	// Replace 2 or more consecutive newlines with a single newline
	doubleNewlineRE := regexp.MustCompile(`\n{2,}`)
	s = doubleNewlineRE.ReplaceAllString(s, "\n")
	return s
}
