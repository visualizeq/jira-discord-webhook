package utils

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type UserMapping map[string]string

var jiraToDiscord UserMapping

func LoadUserMapping(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw struct {
		JiraToDiscord map[string]string `toml:"jira_to_discord"`
	}
	if err := toml.Unmarshal(f, &raw); err != nil {
		return err
	}
	jiraToDiscord = raw.JiraToDiscord
	return nil
}

func DiscordMentionForJiraUser(name string) string {
	if id, ok := jiraToDiscord[name]; ok {
		return "<@" + id + ">"
	}
	return name
}
