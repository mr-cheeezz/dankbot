# DankBot

DankBot is a Twitch-first bot runtime plus a full web dashboard for moderation, commands, modules, integrations, and public profile pages.

## What Is In This Repo

- `cmd/bot`: bot runtime (chat/event handling + modules)
- `cmd/web`: web server (dashboard + public APIs/pages)
- `web/`: React + MUI frontend
- `pkg/`: core packages (modules, postgres stores, twitch/discord clients, web controllers)
- `migrations/`: SQL migrations embedded into the binary
- `configs/`: example/test INI config files

## Requirements

- Go 1.22+ (recommended)
- Node.js 20+ and npm
- PostgreSQL
- Redis (recommended for pub/sub and worker coordination)
- Twitch app credentials (and optional Spotify/Discord/Streamlabs/StreamElements credentials)

## Quick Start

1. Copy and edit config:

```bash
cp configs/example.ini configs/test.ini
```

2. Validate config:

```bash
go run ./cmd/setup --config configs/test.ini
```

3. Run web server:

```bash
go run ./cmd/web --config configs/test.ini
```

4. Run bot worker:

```bash
go run ./cmd/bot --config configs/test.ini
```

5. Frontend dev (optional):

```bash
npm --prefix web install
npm --prefix web run dev
```

## Build

Backend:

```bash
go build ./...
```

Frontend:

```bash
npm --prefix web run build
```

## Config Notes

- Primary app IDs:
  - `main.bot_id`
  - `main.streamer_id`
  - `main.admin_id`
- Main database DSN:
  - `main.db`
- OAuth/Integrations:
  - `twitch.*`
  - `spotify.*`
  - `discord.*`
  - `streamlabs.*`
  - `streamelements.*`
- Public web URL and binds:
  - `web.public_url`
  - `web.bind_addr`

Use `configs/example.ini` as the source of truth for all supported keys.

## Release / Tagging

Create an annotated release tag:

```bash
git tag -a v0.9.0-beta.1 -m "DankBot v0.9.0-beta.1"
git push origin v0.9.0-beta.1
```

If this is your first push of branch changes:

```bash
git push origin <your-branch>
git push origin v0.9.0-beta.1
```

