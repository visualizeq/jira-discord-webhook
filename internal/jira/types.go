package jira

// JiraIssue represents a Jira issue payload
// from webhook events.
type Issue struct {
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

// Comment is a Jira issue comment.
type Comment struct {
	Body   string `json:"body"`
	Author struct {
		DisplayName string `json:"displayName"`
	} `json:"author"`
}

// ChangelogItem represents a single change in an update event.
type ChangelogItem struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}

// Changelog groups a list of changed fields.
type Changelog struct {
	Items []ChangelogItem `json:"items"`
}

// Webhook is the top level structure sent by Jira webhooks.
type Webhook struct {
	Issue     Issue      `json:"issue"`
	Comment   *Comment   `json:"comment,omitempty"`
	Changelog *Changelog `json:"changelog,omitempty"`
}
