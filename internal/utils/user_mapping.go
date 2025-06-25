package utils

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type JiraUserMapping struct {
	AccountID   string `yaml:"accountId"`
	DisplayName string `yaml:"displayName"`
	DiscordID   string `yaml:"discordId"`
}

type UserMapping struct {
	JiraToDiscord []JiraUserMapping `yaml:"jira_to_discord"`
}

var jiraToDiscord UserMapping

func LoadUserMapping(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw UserMapping
	if err := yaml.Unmarshal(f, &raw); err != nil {
		return err
	}
	jiraToDiscord = raw
	return nil
}

func DiscordMentionForJiraUser(key string) string {
	for _, u := range jiraToDiscord.JiraToDiscord {
		if u.AccountID == key || u.DisplayName == key {
			return "<@" + u.DiscordID + ">"
		}
	}
	return key
}

var accountIdPattern = regexp.MustCompile(`\[~accountid:([a-zA-Z0-9:.-]+)\]`)

// ReplaceJiraMentionsWithDiscord replaces all [~accountid:...] in text with Discord mentions.
func ReplaceJiraMentionsWithDiscord(text string) string {
	return accountIdPattern.ReplaceAllStringFunc(text, func(match string) string {
		groups := accountIdPattern.FindStringSubmatch(match)
		if len(groups) == 2 {
			return DiscordMentionForJiraUser(groups[1])
		}
		return match
	})
}

var domainPattern = regexp.MustCompile(`([a-zA-Z0-9-]+\.[a-zA-Z0-9.-]+)`)

// ProtectDomains wraps domain-like patterns in triple backticks if the line is only a domain, otherwise single backticks.
func ProtectDomains(s string) string {
	linkPattern := regexp.MustCompile(`\[[^\]\[]+\|[^\]\[]+\]`) // Jira-style [text|url]
	mdLinkPattern := regexp.MustCompile(`\[[^\]]+\]\([^\)]+\)`) // Markdown [text](url)

	lines := strings.Split(s, "\n")
	inCodeBlock := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Find all link ranges (start, end) in the line
		linkRanges := findAllRanges(line, linkPattern)
		linkRanges = append(linkRanges, findAllRanges(line, mdLinkPattern)...)

		wrapDomain := func(domain string, idx int) string {
			if isInRanges(idx, idx+len(domain), linkRanges) || isInBackticks(line, domain) {
				return domain
			}
			return "`" + domain + "`"
		}

		if domainPattern.MatchString(trimmed) && strings.HasPrefix(trimmed, "*") {
			lines[i] = domainPattern.ReplaceAllStringFunc(line, func(domain string) string {
				idx := strings.Index(line, domain)
				return wrapDomain(domain, idx)
			})
		} else if domainPattern.MatchString(trimmed) && len(trimmed) == len(domainPattern.FindString(trimmed)) {
			// If the whole line is a domain, wrap in triple backticks
			lines[i] = "```" + trimmed + "```"
		} else if domainPattern.MatchString(line) {
			// For inline domains, wrap each domain in single backticks if not already in backticks or a link
			lines[i] = domainPattern.ReplaceAllStringFunc(line, func(domain string) string {
				idx := strings.Index(line, domain)
				return wrapDomain(domain, idx)
			})
		}
	}
	return strings.Join(lines, "\n")
}

// findAllRanges returns a slice of (start, end) pairs for all matches of re in s
func findAllRanges(s string, re *regexp.Regexp) [][2]int {
	matches := re.FindAllStringIndex(s, -1)
	var ranges [][2]int
	for _, m := range matches {
		ranges = append(ranges, [2]int{m[0], m[1]})
	}
	return ranges
}

// isInRanges returns true if [start, end) overlaps any range in ranges
func isInRanges(start, end int, ranges [][2]int) bool {
	for _, r := range ranges {
		if start >= r[0] && end <= r[1] {
			return true
		}
	}
	return false
}

// isInBackticks returns true if the domain is already inside backticks in the line
func isInBackticks(line, domain string) bool {
	idx := strings.Index(line, domain)
	if idx == -1 {
		return false
	}
	before := line[:idx]
	count := strings.Count(before, "`")
	return count%2 == 1
}

// isInMarkdownLink returns true if the domain is inside a Markdown link [text](url) or [text|url]
func isInMarkdownLink(line, domain string) bool {
	idx := strings.Index(line, domain)
	if idx == -1 {
		return false
	}
	// Look for [ ... ]( ...domain... ) or [ ... | ...domain... ]
	openBracket := strings.LastIndex(line[:idx], "[")
	closeBracket := strings.Index(line[idx:], "]")
	if openBracket == -1 || closeBracket == -1 {
		return false
	}
	closeBracket += idx
	// Check for ( or | after ]
	if closeBracket+1 < len(line) && (line[closeBracket+1] == '(' || line[closeBracket+1] == '|') {
		return true
	}
	return false
}
