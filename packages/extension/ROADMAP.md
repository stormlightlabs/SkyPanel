# ROADMAP - "SkyPanel"

## Private / Locally Stored Feeds

- Implement feed execution engine (fetch from sources, apply filters)
- Add FeedBuilder UI for rich feed creation/editing
- Support multiple sources per feed
- Implement author, rate-based, label, and keyword filters
- Add feed preview before saving

## Media

- On-the-fly layout switcher UI (stacked/grid/carousel toggle)
    - Grid layout option for multiple images (2x2 for 4 images, 2x1 for 2)
    - Carousel layout with prev/next navigation

## Collapse Multiple Unread per Follower

- Implement Threaded UI as default so replies are nested under parent posts

## My Profile

- Implement edit profile functionality
- Implement share profile functionality
- Add profile refresh
- Add profile caching

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
