# SkyPanel Extension

Browser extension providing quick access to Bluesky via sidepanel and newtab. Built with WXT + Svelte 5.

**Distribution:** Chrome Web Store, Firefox Add-ons, Edge Add-ons

**Optional Integration:** Can connect to self-hosted SkyPanel server for custom feeds and enhanced features.

## Features

**Timeline & Feeds:**

- View timeline, author feeds, and list feeds
- Computed feeds (mutuals, quiet posters)
- Post grouping with read state tracking
- Infinite scroll with cursor-based pagination

**Social:**

- Profile viewing
- Follow/unfollow users
- Like, repost, reply to posts
- Thread viewing

**Search:**

- Post search with filters
- User search

**Storage:**

- All data stored locally in `chrome.storage.local`
- Custom feed definitions (execution not yet implemented)
- Session persistence
- TTL-based caching for computed feeds

## Development

### Setup

```sh
# Install dependencies (from repo root)
pnpm install

# Start development mode
pnpm dev:extension

# For Firefox
pnpm --filter @skypanel/extension dev:firefox
```

### Building

```sh
# Production build
pnpm build:extension

# Output: packages/extension/.output/
```

### Code Quality

```sh
# Type checking
pnpm --filter @skypanel/extension check

# Linting
pnpm --filter @skypanel/extension lint
```

## Architecture

**Tech Stack:**

- WXT (Web Extension Toolkit)
- Svelte 5 with runes (class-based stores)
- TailwindCSS
- @atproto/api

**Message Passing:**

```sh
UI → BackgroundClient → chrome.runtime.sendMessage() → Background Handler → Service → @atproto/api
```

**Key Directories:**

```sh
src/
├── entrypoints/
│   ├── background/       # Service worker
│   ├── sidepanel/        # Sidepanel UI
│   └── newtab/           # New tab UI
├── lib/
│   ├── components/       # Svelte components
│   ├── state/            # Runes-based stores
│   ├── background/       # Background services
│   ├── storage/          # Storage wrappers
│   └── messaging/        # Message types
```

**State Management:**

- Svelte 5 runes (`$state`, `$derived`)
- Class-based singleton stores

## References

- [Extension ROADMAP](./ROADMAP.md) - Planned features
- [WXT Docs](https://wxt.dev/)
- [Chrome Extensions](https://developer.chrome.com/docs/extensions/)
