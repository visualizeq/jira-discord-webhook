package jira

import (
	"fmt"
	"strings"

	"jira-discord-webhook/internal/discord"
)

const (
	issueColor            = 0x00B0F4
	commentColor          = 0x347433
	changelogColor        = 0xFF6F3C
	commentChangelogColor = 0x5409DA
)

// Capitalize returns s with the first letter upper-cased.
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// ToDiscordMessage converts a Jira webhook payload into a Discord message.
func ToDiscordMessage(w Webhook, baseURL string) discord.WebhookMessage {
	issueURL := ""
	if baseURL != "" {
		issueURL = fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), w.Issue.Key)
	}

	embed := discord.Embed{
		Title: fmt.Sprintf("%s: %s", w.Issue.Key, w.Issue.Fields.Summary),
		URL:   issueURL,
	}
	switch {
	case w.Comment != nil && w.Changelog != nil:
		embed.Color = commentChangelogColor
	case w.Comment != nil:
		embed.Color = commentColor
	case w.Changelog != nil:
		embed.Color = changelogColor
	default:
		embed.Color = issueColor
	}
	embed.Description = w.Issue.Fields.Description

	if w.Comment != nil {
		embed.Description = w.Comment.Body
		embed.Fields = append(embed.Fields, discord.Field{
			Name:   "Comment by",
			Value:  w.Comment.Author.DisplayName,
			Inline: true,
		})
	}

	if w.Changelog != nil && len(w.Changelog.Items) > 0 {
		var changes []string
		for _, item := range w.Changelog.Items {
			if item.FromString == "" && item.ToString == "" {
				continue
			}
			name := Capitalize(item.Field)
			if strings.ToLower(item.Field) == "status" {
				name = "Status"
			}
			if item.FromString == "" {
				changes = append(changes, fmt.Sprintf("%s set to %s", name, item.ToString))
			} else {
				changes = append(changes, fmt.Sprintf("%s: %s → %s", name, item.FromString, item.ToString))
			}
		}
		if len(changes) > 0 {
			embed.Fields = append(embed.Fields, discord.Field{
				Name:  "Changes",
				Value: strings.Join(changes, "\n"),
			})
		}
	}

	embed.Fields = append(embed.Fields, discord.Field{Name: "Priority", Value: w.Issue.Fields.Priority.Name, Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Assignee", Value: w.Issue.Fields.Assignee.DisplayName, Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Status", Value: w.Issue.Fields.Status.Name, Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Type", Value: w.Issue.Fields.Issuetype.Name, Inline: true})

	return discord.WebhookMessage{
		Username: "Jira",
		Embeds:   []discord.Embed{embed},
	}
}
