---
sidebar_position: 6
title: Search
---

# search

Search across the Bluesky graph (actors and posts) or interrogate locally cached feeds. Network-backed searches require a valid session.

```bash
skycli search <users|posts|feeds> <query> [flags]
```

## Subcommands

### users

```bash
skycli search users "<query>" [--limit N] [--cursor token] [--json]
```

- Hits `service.SearchActors` with your query.
- Prints actor handle, display name, DID, a bio excerpt (truncated to 100 characters), and follower stats.
- `--limit`/`--cursor` follow the Bluesky pagination pattern.
- `--json` returns the `SearchActorsOutput` object directly.

### posts

```bash
skycli search posts "<query>" [--limit N] [--cursor token] [--json]
```

- Uses `service.SearchPosts` and formats hits via `ui.DisplayFeed`.
- Supports the same pagination flags and JSON output as the users search.
- Ideal for quick content discovery from the terminal.

### feeds

```bash
skycli search feeds "<query>" [--json]
```

- Operates entirely against the local cache (`feedRepo.List`) and performs a case-insensitive substring match on feed name and source URI.
- `--json` emits the models you can later pipe into `jq`.
- Helpful for spelunking cached feed metadata without hitting the API.

## Sample Output (users)

```text
$ skycli search users "stormlight"
Search Results: stormlight

[1] @stormlightlabs.bsky.social
ℹ   Name: Stormlight Labs
ℹ   DID: did:plc:stormabcd1234
ℹ   Bio: Building tools that keep your Bluesky workflows fast and private.
ℹ   Followers: 1842 | Following: 152 | Posts: 326

[2] @stormlightreader.bsky.social
ℹ   DID: did:plc:reader5678
ℹ   Followers: 289 | Following: 97 | Posts: 54

✓ Found 2 user(s)
ℹ Next cursor: bafyreicursorvalue
```

## Tips

- For bulk discovery use `--json` and pipe to `jq`/`fzf`:
  `skycli search posts "sky panel" -j | jq '.posts[].post.uri'`.
- `search feeds` returns cached models; run `fetch feed` or your ingestion pipeline first to make sure the cache is populated.
