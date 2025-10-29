# ROADMAP - SkyPanel Extension

Browser extension for Bluesky, distributed via Chrome Web Store and Firefox Add-ons.

**Note:** Extension works standalone. Server integration is optional for advanced features.

## Server Integration (Optional)

Users can optionally connect the extension to their self-hosted SkyPanel instance (see `/server/`):

- Configure server URL in extension settings
- Sync authentication between extension and server
- Browse and subscribe to custom feed generators from server
- Manage saved/pinned feeds via server API
- Deep linking to web app routes
- Quick access to server features via sidepanel

**Without Server:** Extension still provides timeline, feeds, computed feeds, search, and social features using Bluesky API directly.

## Media

- On-the-fly layout switcher UI (stacked/grid/carousel toggle)
    - Grid layout option for multiple images (2x2 for 4 images, 2x1 for 2)
    - Carousel layout with prev/next navigation

## My Profile

- Implement edit profile functionality
- Implement share profile functionality

## Search

- Add advanced filters (tags, domain, URL)
- Implement quick-scope shorthands (e.g., `from:@handle`, `#tag`, `site:domain.tld`)
- Add pagination for search results
- Add search result preview in feed cards

## Resilience

- Cursor-aware caching, exponential backoff, and transparent error banners.
- Respect Bluesky rate limits and show user friendly notices + retry guidance.
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
