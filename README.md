
# SkyPanel (cli)

A companion command-line & terminal-UI tool for your Bluesky feed ecosystem:

Manage feeds, view timelines, search posts, publish custom feed algorithms.

## What this does

Fetch feeds (timeline, author, list, custom) and display in terminal or full-screen TUI.

Manage **local feed definitions**: create, list, delete, export.

Search posts with filters: author, tag, language, domain.

Publish your feed definition as a **remote feed generator**:

- Uses the official feed-generator API (compatible with Bluesky’s custom feeds system)

Collapse unread posts by the same author in your feed for a cleaner experience.

## Contributing

1. Please write GoDoc-style comments (`what`, `why`, `how`) for public functions
2. Add unit tests (e.g., feed collapse logic, search logic) (where appropriate)
3. Simple TUI integration tests (where appropriate)

## License

[MIT](https://opensource.org/license/mit)
