package utils

import (
	"os"
	"regexp"

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
