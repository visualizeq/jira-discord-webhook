# Jira Discord Webhook

This project provides a small HTTP server written in Go that receives Jira webhook events and forwards them to a Discord webhook.
The server formats issue updates, comments, and transitions into Discord embeds so you can easily track activity from Jira.

## Features

- **Robust Jira wiki/advanced formatting to Markdown/Discord:**
  - Converts Jira wiki-style links (e.g. `[text|http://example.com]`) to Markdown links for Discord.
  - Supports bold (`+bold+`), italics (`*italic*`), underline (`_underline_`), strikethrough (`-strike-`), monospace/code (`{{code}}`), blockquote (`bq. quote`), and removes color markup.
  - Handles advanced blocks: `{noformat}...{noformat}` (as code), `{panel:title=...}...{panel}` (as Discord-styled block), `{code[:lang]}...{code}` (as fenced code block with language).
  - **Strikethrough formatting is robust:**
    - Hostnames, dates, and similar patterns (e.g. `2025-06-03`, `a-b-c-d-e.abc.com`) are not incorrectly formatted with strikethrough.
    - Only true Jira strikethroughs (e.g. `-strike-`) are converted to Discord's `~~strike~~`.
    - Extensive edge case tests are included for all formatting.
- Handles empty comment bodies gracefully (empty comments will result in empty Discord descriptions).
- Debug logging for incoming Jira payloads and outgoing Discord payloads (set logger to debug level to see raw payloads).
- Comprehensive unit tests for all formatting and handler logic.
- **Jira to Discord user mention mapping:**
  - Supports mapping Jira display names to Discord user IDs using a YAML config file (see `USER_MAPPING_PATH`).
  - When a Jira user matches the mapping, Discord mentions (e.g. `<@123456789>`) are used in notifications.

## Configuration

Set the following environment variables (see `.env.example`):

- `DISCORD_WEBHOOK_URL`: Your Discord webhook URL
- `JIRA_BASE_URL`: Base URL for your Jira instance
- `USER_MAPPING_PATH`: Path to the Jira-to-Discord user mapping YAML file (default: `config/user_mapping.yaml`)
- Other variables for port and color customization

## Docker Compose

To use a custom user mapping file with Docker Compose, add a volume mapping in your `compose.yml`:

```yaml
services:
  jira-discord-webhook:
    # ...existing config...
    environment:
      - USER_MAPPING_PATH=/config/user_mapping.yaml
    volumes:
      - ./config/user_mapping.yaml:/config/user_mapping.yaml:ro
```

This ensures the container uses your local `config/user_mapping.yaml` for user mapping.

## Docker Compose Example

Here is a complete example of a `compose.yml` for this project:

```yaml
services:
  jira-discord-webhook:
    image: ghcr.io/visualizeq/jira-discord-webhook:develop
    environment:
      - DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL}
      - JIRA_BASE_URL=${JIRA_BASE_URL}
      - ISSUE_COLOR=${ISSUE_COLOR-0x00B0F4}
      - COMMENT_COLOR=${COMMENT_COLOR-0x347433}
      - CHANGELOG_COLOR=${CHANGELOG_COLOR-0xFF6F3C}
      - COMMENT_CHANGELOG_COLOR=${COMMENT_CHANGELOG_COLOR-0x5409DA}
      - USER_MAPPING_PATH=/config/user_mapping.yaml
    ports:
      - "8080:8080"
    volumes:
      - ./config/user_mapping.yaml:/config/user_mapping.yaml:ro
    restart: always
```

This configuration ensures your local `config/user_mapping.yaml` is available in the container and all required environment variables are set.

## Building

```bash
go build ./cmd
```

## Running

Set the `DISCORD_WEBHOOK_URL` environment variable to your Discord webhook and start the server:

```bash
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
go run ./cmd
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
docker buildx build --platform linux/amd64,linux/arm64 -t jira-discord-webhook .
```

Run the resulting image by providing the required environment variables:

```bash
docker run -e DISCORD_WEBHOOK_URL=... -p 8080:8080 jira-discord-webhook
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

# Example user mapping (config/user_mapping.yaml):
```yaml
jira_to_discord:
  - accountId: "834295173847200064837294"
    displayName: "Random User1"
    discordId: "235702400604700673"
  - accountId: "927461058372910384756120"
    displayName: "Random User2"
    discordId: "927461058372910384"
```
