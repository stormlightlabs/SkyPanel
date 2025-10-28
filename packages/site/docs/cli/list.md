---
sidebar_position: 5
title: List
---

# list

Inspect cached assets that belong to your account: authored posts and known feeds. Listing works fully offline against the SQLite cache, but the `posts` variant will fetch fresh data if you have a valid session.

```bash
skycli list [subcommand] [flags]
```

`list` defaults to `list posts` when no subcommand is provided.

## Subcommands

### posts

```bash
skycli list posts [--limit N] [--json]
```

- Ensures you are logged in, fetches your DID, and calls `service.GetAuthorFeed` to retrieve your most recent posts.
- Printed output mirrors `fetch timeline`, but scoped to your account.
- `--limit` controls the page size (default 25).
- `--json` returns the raw API payload.

Because this hits the network it is a convenient smoke test to confirm your credentials still work.

### feeds

```bash
skycli list feeds [--refetch] [--json]
```

- Reads from the local feed repository (`feedRepo.List`) and prints feed metadata.
- `--refetch` (`-r`) is wired up but currently prints a warning because `GetUserFeeds` is not implemented yet; expect a local-only view.
- `--json` returns the stored feed models as-is.

Use this view to discover local feed IDs before an export or to audit the cache.

## Sample Output (feeds)

```text
$ skycli list feeds
Your Feeds

[1] Sky News Wire
ℹ   ID: 7f69d974-0e36-4c5d-8538-1d0f1f0f9b44
ℹ   Source: at://did:plc:newsdesk/app.bsky.feed.generator/skypulse
ℹ   Local: false
ℹ   Created: 2024-10-20T19:12:03Z

[2] Personal Highlights
ℹ   ID: 12d53a46-74f1-4b58-b27c-71b62b3a9c90
ℹ   Source: at://did:plc:you/app.bsky.feed.generator/curated
ℹ   Local: true
ℹ   Created: 2024-10-22T03:45:54Z

✓ Total: 2 feed(s)
```

## Tips

- Pair with `fetch feed` to populate cache entries, then `list feeds` to confirm they landed.
- `list posts -j` is handy when you need exact URIs for scripting (`jq '.[0].uri'` etc.).
