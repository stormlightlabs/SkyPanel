---
sidebar_position: 8
title: Export
---

# export

Write cached data to disk for archival, reporting, or downstream tooling. Each export subcommand outputs a timestamped file in the current working directory.

```bash
skycli export <feed|profile|post> <identifier> [flags]
```

## Subcommands

### feed

```bash
skycli export feed <feed-id> [--format json|csv|txt] [--size N]
```

- Looks up the feed in the local cache (`feedRepo.Get`) using a UUID.
- Pulls posts from `postRepo.QueryByFeedID`; only posts already stored locally are exported.
- `--format` (`-f`) defaults to `json`; CSV and TXT are also available.
- `--size` (`-s`) limits the number of posts exported (default 25).
- Writes files named like `feed_<feed-id>_2024-10-27.json`.

If no posts are cached you will see a warning and no file is created—run `fetch feed` first or confirm your ingestion pipeline.

### profile

```bash
skycli export profile <handle-or-did> [--format json|txt]
```

- Fetches the latest profile from the API (`service.GetProfile`), so a valid login is required.
- Supports JSON (full payload) or TXT (formatted summary) output.
- Filenames follow `profile_<handle>_YYYY-MM-DD.<ext>`.

### post

```bash
skycli export post <post-uri-or-bsky-url> [--format json|txt]
```

- Accepts either AT URIs or browser URLs; identifiers are normalized via `parsePostURI`.
- Fetches the post (`service.GetPosts`) and persists the first hit.
- JSON gives you the full `FeedViewPost` (including embeds, labels, etc.), while TXT mirrors the pretty printer used in `view`.

## Sample Output (feed export)

```text
$ skycli export feed 7f69d974-0e36-4c5d-8538-1d0f1f0f9b44 --format csv --size 50
✓ Exported 50 post(s) to feed_7f69d974-0e36-4c5d-8538-1d0f1f0f9b44_2024-10-27.csv
```

## Tips

- After exporting, compress the file explicitly (e.g., `gzip feed_7f69d974-..._2024-10-27.json`) if you want to keep long-term archives small.
- `export post` is handy for sharing a textual snapshot when you cannot rely on the web UI staying available.
- SkyCLI never overwrites existing files; rerunning the same command on a later date produces a new file with the current date suffix.
