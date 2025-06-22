# Jira Discord Webhook

This project provides a small HTTP server written in Go that receives Jira webhook events and forwards them to a Discord webhook.
The server formats issue updates, comments, and transitions into Discord embeds so you can easily track activity from Jira.

## Building

```bash
go build
```

## Running

Set the `DISCORD_WEBHOOK_URL` environment variable to your Discord webhook and start the server:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
go run .
```

Set `JIRA_BASE_URL` to the base URL for your Jira instance so links in Discord messages work correctly:

```bash
export JIRA_BASE_URL="https://your-company.atlassian.net/browse"
```

The server listens on port `8080` by default. You can override this by setting the `PORT` environment variable.

Jira should be configured to send webhooks to `http://your-server:8080/webhook`.

Issue comments will appear in Discord with the comment text and author.
When an issue transitions between statuses, the change will be included in the notification.
If a webhook contains multiple field updates, all of the changes are summarized in a single Discord message so you can see everything that changed at a glance.

## Docker

This repository includes a multi-architecture `Dockerfile`. Build images for multiple platforms with Docker Buildx:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t my/jira-hook .
```

Run the resulting image by providing the required environment variables:

```bash
docker run -e DISCORD_WEBHOOK_URL=... -p 8080:8080 my/jira-hook
```

## Postman Collection

The `postman` directory contains a collection with example webhook requests.
Import `postman/jira-discord-webhook.postman_collection.json` into Postman to
manually trigger the server with sample issue, comment, and changelog payloads.
