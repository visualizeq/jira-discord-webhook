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
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if domainPattern.MatchString(trimmed) && strings.HasPrefix(trimmed, "*") {
			// For bullet lines, wrap domain part in single backticks
			lines[i] = domainPattern.ReplaceAllStringFunc(line, func(domain string) string {
				return "`" + domain + "`"
			})
		} else if domainPattern.MatchString(trimmed) && len(trimmed) == len(domainPattern.FindString(trimmed)) {
			// If the whole line is a domain, wrap in triple backticks
			lines[i] = "```" + trimmed + "```"
		} else if domainPattern.MatchString(line) {
			// For inline domains, wrap each domain in single backticks
			lines[i] = domainPattern.ReplaceAllStringFunc(line, func(domain string) string {
				return "`" + domain + "`"
			})
		}
	}
	return strings.Join(lines, "\n")
}
