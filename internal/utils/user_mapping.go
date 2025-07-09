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
var simpleUserPattern = regexp.MustCompile(`\[~([a-zA-Z0-9:.-]+)\]`)

// ReplaceJiraMentionsWithDiscord replaces all [~accountid:...] and [~user] in text with Discord mentions.
func ReplaceJiraMentionsWithDiscord(text string) string {
	// First handle accountid pattern
	text = accountIdPattern.ReplaceAllStringFunc(text, func(match string) string {
		groups := accountIdPattern.FindStringSubmatch(match)
		if len(groups) == 2 {
			return DiscordMentionForJiraUser(groups[1])
		}
		return match
	})

	// Then handle simple user pattern (only if accountid pattern didn't match)
	text = simpleUserPattern.ReplaceAllStringFunc(text, func(match string) string {
		groups := simpleUserPattern.FindStringSubmatch(match)
		if len(groups) == 2 {
			// Don't convert if it's already a Discord mention
			if strings.HasPrefix(groups[1], "@") {
				return match
			}
			mapped := DiscordMentionForJiraUser(groups[1])
			// If no mapping found, default to @user format
			if mapped == groups[1] {
				return "@" + groups[1]
			}
			return mapped
		}
		return match
	})

	return text
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
