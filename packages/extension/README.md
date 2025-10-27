# SkyPanel

SkyPanel is a Chrome side panel extension that supercharges your Bluesky browsing built with WXT & Svelte.

## Features

- Build feeds from timeline, lists, authors, or saved searches.
    - **Private/local** feeds stored in `chrome.storage.local` (never uploaded).
- **Collapse** multiple unread posts per follow with accurate counts.
- **Defaults:**
    - Mutuals (follows ∩ followers)
    - Quiet Posters (lower post frequency, recent-first)
    - Image only with slideshow view
    - Video only with vertical scroll
- Search with filters (`app.bsky.feed.searchPosts`) and "save as feed."
- MV3 **Side Panel** UI-persistent, fast, and doesn’t reload with page navigations.

## Architecture

- Service Worker: owns `AtpAgent` session, rate-limit/backoff, caching, and message bus to UI
- Side Panel (Svelte): feed browser + composer, search UI, collapsible groups
- Storage: `chrome.storage.local` for sessions (JWT/app password session data) & feed definitions; ephemeral in-memory caches for result pages.
- APIs Used:
    - Timeline: `app.bsky.feed.getTimeline`
    - Author feed: `app.bsky.feed.getAuthorFeed`
    - List feed: `app.bsky.feed.getListFeed`
    - Graph: `app.bsky.graph.getFollows`, `app.bsky.graph.getFollowers`
    - Search: `app.bsky.feed.searchPosts`
    - Public AppView base: `https://public.api.bsky.app` (for public endpoints); authenticated calls proxied via user’s PDS
