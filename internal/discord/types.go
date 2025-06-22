package discord

// WebhookMessage describes the payload sent to Discord webhook.
type WebhookMessage struct {
Username string   `json:"username,omitempty"`
Embeds   []Embed  `json:"embeds"`
}

// Embed represents a Discord embed.
type Embed struct {
Title       string   `json:"title"`
URL         string   `json:"url,omitempty"`
Description string   `json:"description,omitempty"`
Fields      []Field  `json:"fields,omitempty"`
}

// Field represents an embed field.
type Field struct {
Name   string `json:"name"`
Value  string `json:"value"`
Inline bool   `json:"inline"`
}

