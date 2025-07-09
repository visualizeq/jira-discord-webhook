# Jira Discord Webhook

This project provides a small HTTP server written in Go that receives Jira webhook events and forwards them to a Discord webhook.
The server formats issue updates, comments, and transitions into Discord embeds so you can easily track activity from Jira.

## Features

- **Robust Jira wiki/advanced formatting to Markdown/Discord:**
  - Converts Jira wiki-style links (e.g. `[text|http://example.com]`) to Markdown links for Discord.
  - Supports bold (`+bold+`), italics (`*italic*`), underline (`_underline_`), strikethrough (`-strike-`), monospace/code (`{{code}}`), blockquote (`bq. quote`), and removes color markup.
  - Handles advanced blocks: `{noformat}...{noformat}` (as code), `{panel:title=...}...{panel}` (as Discord-styled block), `{code[:lang]}...{code}` (as fenced code block with language).
  - **Jira image markup conversion:** Converts Jira image markup like `!Screenshot 2025-04-17 at 14.39.17.png|...!` to `Screenshot 2025-04-17 at 14.39.17.png` (wrapped in backticks) for Discord.
  - **Strikethrough formatting is robust:**
    - Hostnames, dates, and similar patterns (e.g. `2025-06-03`, `a-b-c-d-e.abc.com`) are not incorrectly formatted with strikethrough.
    - Only true Jira strikethroughs (e.g. `-strike-`) are converted to Discord's `~~strike~~`.
    - Extensive edge case tests are included for all formatting.
- **Advanced domain and filename protection:**
  - All domain-like and filename-like patterns in messages are automatically wrapped in backticks (inline code) for Discord.
  - Prevents unwanted Markdown formatting while preserving readability.
  - Handles complex domains like `a-b-c-d-e.abc.com` as single units.
  - Filenames with spaces and multiple extensions are properly protected.
  - Excludes domains already in URLs or Markdown links to avoid double-wrapping.
- Handles empty comment bodies gracefully (empty comments will result in empty Discord descriptions).
- Debug logging for incoming Jira payloads and outgoing Discord payloads (set logger to debug level to see raw payloads).
- Comprehensive unit tests for all formatting and handler logic.
- **Jira to Discord user mention mapping:**
  - Supports mapping Jira account IDs and display names to Discord user IDs using a YAML config file (see `USER_MAPPING_PATH`).
  - Handles both `[~accountid:...]` and `[~user]` mention formats from Jira.
  - When a Jira user matches the mapping, Discord mentions (e.g. `<@123456789>`) are used in notifications.
  - Falls back to `@username` format for unmapped users.

## User Mapping Configuration

The webhook supports mapping Jira users to Discord users for proper mentions. Create a YAML file (default: `config/user_mapping.yaml`) with the following structure:

```yaml
jira_to_discord:
  - accountId: "834295173847200064837294"
    displayName: "Random User1"
    discordId: "235702400604700673"
  - accountId: "927461058372910384756120"
    displayName: "Random User2"
    discordId: "927461058372910384"
```

The webhook will:
- Convert `[~accountid:712020:1a0df378-f399-40bb-bc8f-ca66d95e68d7]` to `<@123456789012345678>`
- Convert `[~Random User1]` to `<@235702400604700673>`
- Fall back to `@username` for unmapped users

## Configuration

Set the following environment variables (see `.env.example`):

- `DISCORD_WEBHOOK_URL`: Your Discord webhook URL
- `JIRA_BASE_URL`: Base URL for your Jira instance
- `USER_MAPPING_PATH`: Path to the Jira-to-Discord user mapping YAML file (default: `config/user_mapping.yaml`)
- `PORT`: Server port (default: 8080)
- Color customization variables for different event types

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

The project includes comprehensive unit tests for all functionality:

- **Jira-to-Markdown conversion tests:** All formatting patterns, edge cases, and image markup conversion
- **User mapping tests:** Both account ID and display name mappings, fallback behavior
- **Domain and filename protection tests:** Complex domains, filenames with spaces, edge cases
- **Handler tests:** Webhook processing, error handling, payload validation

For summarized test output install [tparse](https://github.com/mfridman/tparse)
and run:

```bash
go install github.com/mfridman/tparse@latest
go test -json ./... | tparse -all
```

Run specific test suites:

```bash
# Test user mapping functionality
go test -v ./internal/utils -run TestReplaceJiraMentionsWithDiscord

# Test Jira-to-Markdown conversion
go test -v ./internal/jira -run TestJiraToMarkdown

# Test domain and filename protection
go test -v ./internal/utils -run TestProtectDomainsAndFiles
```

## Releases

This project automatically generates release notes using [git-cliff](https://github.com/orhun/git-cliff) whenever changes are pushed to the `main` branch or a tag is created.

## Advanced Features

### Jira Image Markup Conversion

The webhook automatically converts Jira image markup to Discord-friendly format:

- `!Screenshot 2025-04-17 at 14.39.17.png|width=681,height=552!` → `Screenshot 2025-04-17 at 14.39.17.png`
- Handles filenames with spaces and complex attributes
- Wraps the filename in backticks for proper Discord formatting

### Domain and Filename Protection

- All domain-like and filename-like patterns in messages are automatically wrapped in backticks (inline code) for Discord, except when part of a Markdown/Jira link or image/attachment.
- Filenames with double underscores are normalized (e.g., `move__bank__cus_mapping.sh` → `move_bank_cus_mapping.sh`) and wrapped in backticks.
- Bare domains (e.g., `a-b-c-d-e.abc.com`) are always wrapped in backticks for clarity and to prevent unwanted Markdown formatting.
- Complex domains with multiple hyphens are handled as single units to prevent splitting
- Extensive unit tests ensure that Markdown formatting is robust and Discord-friendly, including edge cases for domains, filenames, and links.

### User Mention Mapping

The webhook supports sophisticated user mention mapping:

- **Account ID mapping:** Maps Jira account IDs (e.g., `712020:1a0df378-f399-40bb-bc8f-ca66d95e68d7`) to Discord user IDs
- **Display name mapping:** Maps Jira display names to Discord user IDs for backwards compatibility
- **Multiple mention formats:** Handles both `[~accountid:...]` and `[~user]` patterns from Jira
- **Fallback behavior:** Uses `@username` format for unmapped users
- **Real-time processing:** User mappings are applied during webhook processing without additional API calls

## Project Structure

```
├── cmd/                     # Main application entry point
├── internal/
│   ├── discord/            # Discord webhook client and types
│   ├── handler/            # HTTP webhook handler
│   ├── jira/               # Jira payload processing and conversion
│   │   ├── jira2md.go      # Jira markup to Markdown conversion
│   │   ├── convert.go      # Jira to Discord message conversion
│   │   └── testdata/       # Test payloads for various scenarios
│   └── utils/              # Utility functions
│       ├── user_mapping.go # Jira to Discord user mapping
│       └── helper.go       # Domain and filename protection
├── config/                 # Configuration files
│   └── user_mapping.yaml   # User mapping configuration
├── postman/                # Postman collection for testing
└── logs/                   # Application logs
```

Key files:
- `internal/jira/jira2md.go`: Converts Jira wiki markup to Markdown
- `internal/utils/user_mapping.go`: Handles user mention mapping
- `internal/utils/helper.go`: Protects domains and filenames
- `config/user_mapping.yaml`: User mapping configuration
