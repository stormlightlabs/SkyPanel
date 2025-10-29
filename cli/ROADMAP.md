# ROADMAP - "SkyCLI"

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

CLI commands for managing feeds on SkyFeed server (see `/server/ROADMAP.md`):

- `skycli feed create` - Create new feed definition on server
- `skycli feed list` - List user's feeds
- `skycli feed edit` - Update feed definition
- `skycli feed delete` - Delete feed
- `skycli feed test` - Test feed algorithm locally with sample data
- `skycli feed export/import` - Share feed definitions as JSON

Done?: CLI can create, list, edit, delete, and test feed definitions via server API.

## Collapsing unread logic & filter/search support

- In TUI, support collapsed unread by author logic (as in extension).
- Support advanced search: `search posts` with filters (author, tag, lang, domain).
- Done?: TUI or CLI supports search returning results; collapse unread groups.

## Feed publishing to Bluesky

CLI commands for publishing feeds to AT Protocol (requires SkyFeed server):

- `skycli feed publish <feedID>` - Publish feed to user's Bluesky account
- `skycli feed unpublish <feedID>` - Remove published feed
- `skycli feed share <feedID>` - Generate shareable at:// URI

Publishing process:

1. Feed definition sent to SkyFeed server
2. Server creates feed generator record in user's repository
3. Feed becomes discoverable via at:// URI
4. Users can subscribe in any Bluesky client

Done?: CLI publishes feeds via server, outputs at:// URI for sharing.

## Server Management

CLI commands for managing the SkyFeed server (see `/server/ROADMAP.md` for server architecture):

- `skycli server init` - Initialize new feed generator service
- `skycli server start` - Start local development server
- `skycli server stop` - Stop running server
- `skycli server deploy` - Deploy to configured platform (Fly.io, Docker, etc.)
- `skycli server status` - Check service health and stats
- `skycli server logs` - View server logs
- `skycli server config` - Manage server configuration

Done?: CLI can initialize, start, deploy, and manage SkyFeed server instances.

## Packaging & distribution

- Build binaries for major OSes (Windows, macOS, Linux).
- Write unit tests (feed logic, collapse logic) and TUI smoke tests.
- Release via GitHub, include versioning.
    Done?: CI pipeline passes, binaries available.
