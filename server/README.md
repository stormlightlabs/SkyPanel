# SkyPanel Server

Self-hosted Bluesky web client and custom feed generator.

**Status:** In planning and development. See [ROADMAP.md](./ROADMAP.md) for implementation plan.

## Overview

SkyPanel Server will be a self-hosted service (like Forgejo/Gitea) that provides:

- Full-featured Bluesky web client
- Custom feed generator following AT Protocol
- Single binary deployment with embedded web app
- SQLite by default (PostgreSQL optional)
- Optional integration with SkyPanel browser extension

**Distribution:** Self-deployed on user's infrastructure via Docker or binary

## Architecture

**Tech Stack:**

- Go 1.24+ for backend
- Svelte 5 + SvelteKit for web app
- SQLite (default) or PostgreSQL
- Standard library HTTP server

**Components:**

- Web application (Svelte 5 SPA)
- Feed generator service (AT Protocol)
- REST API for extension integration
- Database layer (SQLite/PostgreSQL)

See [ROADMAP.md](./ROADMAP.md) for detailed implementation plan.

## Development

Not yet implemented. See [ROADMAP.md](./ROADMAP.md) for development phases.

## References

- [ROADMAP.md](./ROADMAP.md) - Implementation plan
- [AT Protocol Docs](https://docs.bsky.app/)
- [Custom Feeds Starter Template](https://docs.bsky.app/docs/starter-templates/custom-feeds)
