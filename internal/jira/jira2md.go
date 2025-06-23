package jira

import (
	"regexp"
)

// JiraToMarkdown converts Jira wiki-style links to Markdown links.
// Example: [text|http://example.com] => [text](http://example.com)
func JiraToMarkdown(s string) string {
	jiraLinkRE := regexp.MustCompile(`\[(.+?)\|([^\]]+)\]`)
	return jiraLinkRE.ReplaceAllStringFunc(s, func(m string) string {
		parts := jiraLinkRE.FindStringSubmatch(m)
		if len(parts) == 3 {
			return "[" + parts[1] + "](" + parts[2] + ")"
		}
		return m
	})
}
