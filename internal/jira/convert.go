package jira

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"jira-discord-webhook/internal/discord"
	"jira-discord-webhook/internal/utils"
)

const (
	issueColor            = 0x00B0F4
	commentColor          = 0x347433
	changelogColor        = 0xFF6F3C
	commentChangelogColor = 0x5409DA
)

// colorFromEnv returns the color defined in the given environment variable. If
// the variable is empty or invalid, def is returned.
func colorFromEnv(name string, def int) int {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	val = strings.TrimPrefix(val, "#")
	v, err := strconv.ParseInt(val, 0, 32)
	if err != nil {
		return def
	}
	return int(v)
}

// Capitalize returns s with the first letter upper-cased.
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// truncateString ensures a string does not exceed max length.
func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// ToDiscordMessage converts a Jira webhook payload into a Discord message.
func ToDiscordMessage(w Webhook, baseURL string) discord.WebhookMessage {
	issueURL := ""
	if baseURL != "" {
		issueURL = fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), w.Issue.Key)
	}

	// Discord embed limits
	const (
		titleMax      = 256
		descMax       = 4096
		fieldNameMax  = 256
		fieldValueMax = 1024
		maxFields     = 25
	)

	title := truncateString(fmt.Sprintf("%s: %s", w.Issue.Key, w.Issue.Fields.Summary), titleMax)
	var desc string
	if w.Comment != nil {
		desc = ""
	} else {
		desc = w.Issue.Fields.Description
		desc = utils.ProtectDomains(desc)
		desc = utils.ReplaceJiraMentionsWithDiscord(desc)
		desc = JiraToMarkdown(desc)
		desc = truncateString(desc, descMax)
	}

	embed := discord.Embed{
		Title:       title,
		URL:         issueURL,
		Description: "",
	}
	switch {
	case w.Comment != nil && w.Changelog != nil:
		embed.Color = colorFromEnv("COMMENT_CHANGELOG_COLOR", commentChangelogColor)
	case w.Comment != nil:
		embed.Color = colorFromEnv("COMMENT_COLOR", commentColor)
	case w.Changelog != nil:
		embed.Color = colorFromEnv("CHANGELOG_COLOR", changelogColor)
	default:
		embed.Color = colorFromEnv("ISSUE_COLOR", issueColor)
	}

	// Add Description as a separate field if present
	if desc != "" {
		embed.Fields = append(embed.Fields, discord.Field{
			Name:   truncateString("Description", fieldNameMax),
			Value:  truncateString(desc, fieldValueMax),
			Inline: false,
		})
	}

	if w.Comment != nil {
		commentBody := w.Comment.Body
		commentBody = utils.ProtectDomains(commentBody)
		fmt.Println("[DEBUG] After ProtectDomains:", commentBody)
		commentBody = utils.ReplaceJiraMentionsWithDiscord(commentBody)
		fmt.Println("[DEBUG] After ReplaceJiraMentionsWithDiscord:", commentBody)
		commentBody = JiraToMarkdown(commentBody)
		fmt.Println("[DEBUG] After JiraToMarkdown:", commentBody)
		commentBody = truncateString(commentBody, fieldValueMax)
		embed.Fields = append(embed.Fields, discord.Field{
			Name:   truncateString("Comment", fieldNameMax),
			Value:  commentBody,
			Inline: false,
		})
		embed.Fields = append(embed.Fields, discord.Field{
			Name:   truncateString("Comment by", fieldNameMax),
			Value:  truncateString(utils.DiscordMentionForJiraUser(w.Comment.Author.DisplayName), fieldValueMax),
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
			var change string
			if item.FromString == "" {
				change = fmt.Sprintf("%s set to %s", name, utils.DiscordMentionForJiraUser(item.ToString))
			} else {
				change = fmt.Sprintf("%s: %s → %s", name, utils.DiscordMentionForJiraUser(item.FromString), utils.DiscordMentionForJiraUser(item.ToString))
			}
			changes = append(changes, truncateString(change, fieldValueMax))
		}
		if len(changes) > 0 {
			field := discord.Field{
				Name:  truncateString("Changes", fieldNameMax),
				Value: truncateString(strings.Join(changes, "\n"), fieldValueMax),
			}
			embed.Fields = append(embed.Fields, field)
		}
	}

	// Inline fields: show as plain text, no markdown link
	embed.Fields = append(embed.Fields, discord.Field{Name: "Priority", Value: truncateString(w.Issue.Fields.Priority.Name, fieldValueMax), Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Assignee", Value: truncateString(utils.DiscordMentionForJiraUser(w.Issue.Fields.Assignee.DisplayName), fieldValueMax), Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Status", Value: truncateString(w.Issue.Fields.Status.Name, fieldValueMax), Inline: true})
	embed.Fields = append(embed.Fields, discord.Field{Name: "Type", Value: truncateString(w.Issue.Fields.Issuetype.Name, fieldValueMax), Inline: true})

	// Discord allows max 25 fields
	if len(embed.Fields) > maxFields {
		embed.Fields = embed.Fields[:maxFields]
	}

	return discord.WebhookMessage{
		Username: "Jira",
		Embeds:   []discord.Embed{embed},
	}
}
