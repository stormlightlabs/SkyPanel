---
sidebar_position: 7
title: View
---

# view

Render Bluesky content with rich, human-readable formatting. Unlike `fetch`, the `view` subcommands target specific resources and apply extra parsing (e.g., bsky.app URLs).

```bash
skycli view <feed|post|profile> <identifier> [flags]
```

All variants expect an authenticated session.

## Subcommands

### feed

```bash
skycli view feed <feed-uri-or-local-id> [--limit N] [--cursor token] [--json]
```

- Mirrors `fetch feed` but focuses on inspection, printing the feed URI as the title.
- Accepts either an AT URI or a cached feed UUID (resolved through `feedRepo`).
- Useful for sanity-checking a generator before exporting it.

### post

```bash
skycli view post <post-uri-or-bsky-url> [--json]
```

- Accepts AT URIs (`at://did:.../app.bsky.feed.post/<rkey>`) or full `https://bsky.app/profile/<handle>/post/<rkey>` URLs.
- Converts URLs to URIs via `parsePostIdentifier`, fetches the record with `service.GetPosts`, and prints it using `ui.DisplayFeed`.
- `--json` returns the `FeedViewPost` object if you need to inspect embeds or facets programmatically.

### profile

```bash
skycli view profile <handle-or-did> [--with-posts] [--json]
```

- Retrieves the profile via `service.GetProfile` and displays a header containing handle, display name, bio, and follower counts.
- With `--with-posts` (`-p`), fetches the latest 10 posts and prints them beneath the profile header.
- `--json` returns the `ActorProfile` raw JSON.

## Sample Output (profile with posts)

```text
$ skycli view profile @stormlightlabs.bsky.social --with-posts
@stormlightlabs.bsky.social
Stormlight Labs
  Building tools that keep your Bluesky workflows fast and private.
ℹ   Followers: 1842 | Following: 152 | Posts: 326

Recent Posts
[1] Post by @stormlightlabs.bsky.social
ℹ   URI: at://did:plc:stormabcd1234/app.bsky.feed.post/3kf4b2q
  New CLI docs just landed. Power users, this one's for you.
ℹ   ❤️  67 | 🔁 11 | 💬 9
ℹ   Indexed: 2024-10-27T19:04:11Z

[2] Post by @stormlightlabs.bsky.social
ℹ   URI: at://did:plc:stormabcd1234/app.bsky.feed.post/3kf46xf
  Shipping a feed export tweak: CSV now includes language hints.
ℹ   ❤️  48 | 🔁 6 | 💬 3
ℹ   Indexed: 2024-10-27T15:26:30Z

✓ Showing 2 post(s)
```

## Tips

- Pair with `fzf` for quick browsing: `skycli list feeds | rg '^ℹ   ID' | awk '{print $3}' | fzf | xargs skycli view feed`.
- `view post` is perfect for pasting a bsky.app link from the browser and still staying in the terminal.
