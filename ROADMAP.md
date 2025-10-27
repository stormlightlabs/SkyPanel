# ROADMAP - "SkyCLI"

- Basic commands: `login`, `list`:`list feeds`, `view`: `feed <feedURI>`, `export`: `feed <localFeedID>`.

## Agent Backend integration

- Command Groups: `fetch`, `search`, `view`
- Commands:
    - `fetch`
        1. `timeline` - default
        2. `feed <feedURI|localFeedID>`
        3. `author <actor>`
    - `view`
        1. `view post <postID|postURL>`
        2. `feed <feedURI|localFeedID>`
        3. `profile <actor>`
    - `search`
        1. `search feeds <query>`
        2. `search posts <query>`
        3. `search users <query>`
    - `list` (refetches with flag `-r`)
        - `posts`: fetches user's own posts and shows them - default
        - `feeds`: fetches user's own feeds and shows them
    - `export` (keep existing command signature)
        - `feed <localFeedID>`
        - `profile <actor>`
        - `post <postID|postURL>`

- Done?: CLI can fetch and display the posts for timeline or author.

## TUI view with bubbletea

- Implement interactive feed-browser (pager) in terminal (scrolling list of posts, expand post details).
- Command `tui` launches full-screen UI; allow selection of feed, filtering, unread collapse.
- Done?: TUI runs, shows feed, allows basic navigation (up/down).

## Save & Archive System

- Unified Save Interface
    - Introduce `save` domain layer handling all save operations (`SaveLink`, `SavePost`, `SaveMedia`).
    - Add a CLI command family
- Save Links
    - Fetch OpenGraph metadata (title, description, image, favicon).
    - Store link metadata in JSON (`links.json`) and cache preview thumbnails.
    - Allow tagging (`--tags`) and searching by tag in TUI.
- Save Posts (Account)
    - Integrate with Bluesky API (`app.bsky.graph.like` or custom record type) to save posts to the user’s account.
    - Mirror saved post metadata locally for offline viewing.
    - Provide sync command `skycli sync saved`.
- Save Posts (Local Markdown)
    - Export selected posts as Markdown files to:

    ```sh
    ~/Library/Application Support/SkyCLI/saved/posts/
    ~/.local/share/skycli/saved/posts/
    ```

    - Markdown includes:
        - Author handle and timestamp
        - Post text, mentions, and embedded links
        - YAML frontmatter (tags, source URL, Bluesky URI)
- Download Images or Video
    - Parse attached media from post records and download via HTTP with correct filenames.
    - Save to `/media/images/` and `/media/videos/` subdirectories.
    - Maintain reference paths in Markdown frontmatter.

### Done?

- `skycli save` commands handle links, posts, and media
- Local Markdown and media files stored in per-user app data directory
- Optional remote sync for account-level saved posts
- "Saved Items" TUI panel displays and filters all saved entries
- Repository and service layers fully abstract storage and network APIs

## Feed definition management

- Manage local feed definitions in CLI: `feed create`, `feed delete`, `export feed` (to remote generator), `import feed`.
- Persist definitions in a local database or config (e.g., BoltDB or simple JSON file).
- Done?: CLI can create, list, delete feed definitions.

## Collapsing unread logic & filter/search support

- In TUI, support collapsed unread by author logic (as in extension).
- Support advanced search: `search posts` with filters (author, tag, lang, domain).
- Done?: TUI or CLI supports search returning results; collapse unread groups.

## Remote generator publishing

- CLI command: `feed publish <localFeedID> --service-endpoint <url>` which uses the feed generator starter kit spec to:
    1. register the feed algorithm
    2. declare its metadata, and publish it.
- Use template[^bsky-docs]
- Done?: CLI outputs the generated feed URI that can be used in Bluesky.

## Feed Generator

A feed generator in Bluesky terms is just an HTTP service implementing the AT Protocol feed endpoints, e.g.:

- `app.bsky.feed.describeFeedGenerator`
- `app.bsky.feed.getFeedSkeleton`

When Bluesky clients query these, the generator responds with JSON listing post URIs and optional metadata.

You can spin this up locally, on localhost:PORT, and register it either:

- Temporarily (for testing with your own Bluesky account)
- Permanently (if you later host it and assign a DID)

## Packaging & distribution

- Build binaries for major OSes (Windows, macOS, Linux).
- Write unit tests (feed logic, collapse logic) and TUI smoke tests.
- Release via GitHub, include versioning.
    Done?: CI pipeline passes, binaries available.

[^bsky-docs]: <https://docs.bsky.app/docs/starter-templates/custom-feeds> "Custom Feeds | Bluesky"
