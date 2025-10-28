# ROADMAP - "SkyPanel"

## Private / Locally Stored Feeds

- Local schema for "feed definitions" (JSON): sources (author/list/timeline), include/exclude sets, rate-based rules, label/mute filters.
    - Persist with `chrome.storage.local`; version with simple migrations.
- No remote publishing; these feeds are **private** by design.
- Done?: User can create, rename, clone, delete local feeds; all entirely client-side.

## Media

- Download videos
- Download images

## Collapse Multiple Unread per Follower

- Track per-actor last seen timestamps in a feed context.
    - Group consecutive unread posts by the same author into a collapsible block showing count + preview.
- Mark-read semantics: collapse disappears when user expands or scrolls past.
- Done?: Long runs from prolific authors collapse to a single row with accurate unread counts.

## Search

- UI for `app.bsky.feed.searchPosts` with filters (query string, author, tag/hashtag, domain, lang, date ranges).
    - Add quick-scope shorthands (e.g., `from:@handle`, `#tag`, `site:domain.tld`).
- Paginate and allow "save as local feed" from a search query.
- Done?: Search page returns results with filters; any search can be stored as a private feed.

## Resilience

- Cursor-aware caching, exponential backoff, and transparent error banners.
- Respect Bluesky rate limits and show user friendly notices + retry guidance.
- Done?: Rate-limit events are handled gracefully.

## Finishing Pass

- Accessibility sweep
- Keyboard navigation across lists
- ARIA for collapsibles.
- Release packaging with WXT publish flow and docs.

## References

<https://developer.chrome.com/docs/extensions/develop/concepts/service-workers/basics> "Extension service worker basics - Chrome for Developers"
<https://developer.chrome.com/docs/extensions/reference/api/sidePanel> "chrome.sidePanel | API - Chrome for Developers"
<https://www.npmjs.com/package/%40atproto/api> "atproto/api"
<https://docs.bsky.app/docs/category/http-reference> "HTTP Reference | Bluesky"
<https://docs.bsky.app/docs/starter-templates/custom-feeds> "Custom Feeds"
<https://docs.bsky.app/docs/api/app-bsky-feed-get-timeline> "app.bsky.feed.getTimeline"
<https://docs.bsky.app/docs/api/app-bsky-feed-search-posts> "app.bsky.feed.searchPosts"
<https://docs.bsky.app/docs/advanced-guides/rate-limits> "Rate Limits"
<https://docs.bsky.app/docs/api/app-bsky-graph-get-follows> "app.bsky.graph.getFollows | Bluesky"
<https://docs.bsky.app/docs/api/app-bsky-feed-get-author-feed> "app.bsky.feed.getAuthorFeed"
