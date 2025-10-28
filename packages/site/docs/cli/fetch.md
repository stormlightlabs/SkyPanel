---
sidebar_position: 4
title: Fetch
---

# fetch

Pull posts from Bluesky and print them in a terminal-friendly format. Fetch operations also hydrate the local cache (`cache.db`), which powers later `list`, `search feeds`, and `export` commands.

```bash
skycli fetch [subcommand] [flags]
```

With no subcommand, `fetch` defaults to `fetch timeline`.

## Shared Flags

| Flag | Alias | Description |
| --- | --- | --- |
| `--limit <n>` | `-l` | Page size (default 25). |
| `--cursor <token>` | `-c` | Pagination cursor from the previous call. |
| `--json` | `-j` | Emit the raw JSON payload instead of formatted text. |

All fetch variants require a valid login; they will exit with “not authenticated” if no session is present.

## Subcommands

### timeline

```bash
skycli fetch timeline [--limit N] [--cursor token] [--json]
```

Retrieves the authenticated user's home timeline via `service.GetTimeline`. Each post is printed with author handle, text sample (truncated to 200 characters), reaction counts, and indexed timestamp. When `--json` is set the command returns the API object unmodified—including the feed array and next cursor.

### feed

```bash
skycli fetch feed <feed-uri-or-local-id> [--limit N] [--cursor token] [--json]
```

Accepts either:

- an AT URI such as `at://did:plc:xyz/app.bsky.feed.generator/abc123`, or
- a UUID for a locally stored feed (resolves via `feedRepo.Get` and replaces it with the feed's source URI).

Posts are fetched live from the API (`service.GetAuthorFeed`) and cached. This is how you hydrate the local store before exporting a feed.

### author

```bash
skycli fetch author <handle-or-did> [--limit N] [--cursor token] [--json]
```

Fetches an author's posts and prints a profile header before the timeline. Profiles are cached in SQLite for one hour; if a cached entry is fresh it is reused, otherwise the API is queried and the cache is updated. Useful for quick stalking without leaving the terminal.

## Sample Output (timeline)

```text
$ skycli fetch timeline --limit 3
[1] Post by @sandrakesler.bsky.social
ℹ   URI: at://did:plc:abcd1234/app.bsky.feed.post/3kf46v2
  Quick status update from the terminal-only account. Shipping new feed today!
ℹ   ❤️  18 | 🔁 2 | 💬 5
ℹ   Indexed: 2024-10-27T18:45:03Z

[2] Post by @gardener.bsky.social
ℹ   URI: at://did:plc:wxyz0987/app.bsky.feed.post/3kf3vsa
  Morning photo drop 📷🌿
ℹ   ❤️  102 | 🔁 9 | 💬 12
ℹ   Indexed: 2024-10-27T18:42:11Z

[3] Post by @devfeed.bsky.social
ℹ   URI: at://did:plc:jklo5678/app.bsky.feed.post/3kf3sxt
  Changelog: added audio attachment previews and improved cursor paging.
ℹ   ❤️  37 | 🔁 4 | 💬 3
ℹ   Indexed: 2024-10-27T18:39:27Z

✓ Showing 3 post(s)
ℹ Next cursor: bafyreicursorvalue
```

If you pass `--json` the output switches to a prettified JSON document produced by `ui.DisplayJSON`.

## Tips

- Combine with `jq` for automation: `skycli fetch timeline -j | jq '.feed[].post.author.handle'`.
- Capture cursors and loop for archival pulls; SkyCLI does not auto-paginate to avoid rate spikes.
- The feed and author variants populate the same cache tables used by `export feed`, so fetch first, export later.
