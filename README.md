# Jira Discord Webhook

This project provides a small HTTP server written in Go that receives Jira webhook events and forwards them to a Discord webhook.
The server formats issue updates, comments, and transitions into Discord embeds so you can easily track activity from Jira.

## Building

```bash
go build
```

## Running

Set the `DISCORD_WEBHOOK_URL` environment variable to your Discord webhook and start the server. You can export the variables in your shell or create a `.env` file:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
go run .
```

Copy `.env.example` to `.env` to provide the required variables without exporting them manually.

Set `JIRA_BASE_URL` to the base URL for your Jira instance so links in Discord messages work correctly:

```bash
export JIRA_BASE_URL="https://your-company.atlassian.net/browse"
```

Environment variables from a `.env` file are loaded automatically when the server starts.

The server listens on port `8080` by default. You can override this by setting the `PORT` environment variable.

Jira should be configured to send webhooks to `http://your-server:8080/webhook`.

Issue comments will appear in Discord with the comment text and author.
When an issue transitions between statuses, the change will be included in the notification.
If a webhook contains multiple field updates, all of the changes are summarized in a single Discord message so you can see everything that changed at a glance.
Each message type uses a different embed color so you can quickly see what kind of update occurred.

* Issue events are blue (`#00B0F4`)
* Comment events are green (`#347433`)
* Changelog events are orange (`#FF6F3C`)
* Combined comment and changelog events are purple (`#5409DA`)

You can override these defaults by setting the following environment variables:

```
ISSUE_COLOR=0x00B0F4
COMMENT_COLOR=0x347433
CHANGELOG_COLOR=0xFF6F3C
COMMENT_CHANGELOG_COLOR=0x5409DA
```

Values may be specified in decimal or hexadecimal (with `0x` or `#` prefixes).

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

## Testing

For summarized test output install [tparse](https://github.com/mfridman/tparse)
and run:

```bash
go install github.com/mfridman/tparse@latest
go test -json ./... | tparse -all
```

## Releases

This project automatically generates release notes using [git-cliff](https://github.com/orhun/git-cliff) whenever changes are pushed to the `main` branch or a tag is created.
