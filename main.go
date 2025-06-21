package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Priority    struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Issuetype struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

type JiraComment struct {
	Body   string `json:"body"`
	Author struct {
		DisplayName string `json:"displayName"`
	} `json:"author"`
}

type JiraChangelogItem struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}

type JiraChangelog struct {
	Items []JiraChangelogItem `json:"items"`
}

type JiraWebhook struct {
	Issue     JiraIssue      `json:"issue"`
	Comment   *JiraComment   `json:"comment,omitempty"`
	Changelog *JiraChangelog `json:"changelog,omitempty"`
}

type DiscordWebhookMessage struct {
	Username string         `json:"username,omitempty"`
	Embeds   []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	URL         string         `json:"url,omitempty"`
	Description string         `json:"description,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func sendToDiscord(msg DiscordWebhookMessage) error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL not set")
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload JiraWebhook
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Println("failed to decode jira payload:", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	baseURL := os.Getenv("JIRA_BASE_URL")
	issueURL := ""
	if baseURL != "" {
		issueURL = fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), payload.Issue.Key)
	}

	embed := DiscordEmbed{
		Title: fmt.Sprintf("%s: %s", payload.Issue.Key, payload.Issue.Fields.Summary),
		URL:   issueURL,
	}

	// Use issue description by default
	embed.Description = payload.Issue.Fields.Description

	if payload.Comment != nil {
		embed.Description = payload.Comment.Body
		embed.Fields = append(embed.Fields, DiscordField{
			Name:   "Comment by",
			Value:  payload.Comment.Author.DisplayName,
			Inline: true,
		})
	}

	if payload.Changelog != nil {
		for _, item := range payload.Changelog.Items {
			if strings.ToLower(item.Field) == "status" {
				embed.Fields = append(embed.Fields, DiscordField{
					Name:  "Transition",
					Value: fmt.Sprintf("%s → %s", item.FromString, item.ToString),
				})
				break
			}
		}
	}

	embed.Fields = append(embed.Fields, DiscordField{Name: "Priority", Value: payload.Issue.Fields.Priority.Name, Inline: true})
	embed.Fields = append(embed.Fields, DiscordField{Name: "Assignee", Value: payload.Issue.Fields.Assignee.DisplayName, Inline: true})
	embed.Fields = append(embed.Fields, DiscordField{Name: "Status", Value: payload.Issue.Fields.Status.Name, Inline: true})
	embed.Fields = append(embed.Fields, DiscordField{Name: "Type", Value: payload.Issue.Fields.Issuetype.Name, Inline: true})

	msg := DiscordWebhookMessage{
		Username: "Jira",
		Embeds:   []DiscordEmbed{embed},
	}

	if err := sendToDiscord(msg); err != nil {
		log.Println("failed to send to discord:", err)
		http.Error(w, "failed to send to discord", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
