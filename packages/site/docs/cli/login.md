---
sidebar_position: 2
title: Login
---

# login

Authenticate SkyCLI against Bluesky and persist the encrypted session. A valid session unlocks all network operations (fetch, search, view, export).

```bash
skycli login [--file path] [--handle @name] [--password app-password]
```

Provide either:

- `--file` / `-f`: path to a dotenv-style file containing `BLUESKY_HANDLE` and `BLUESKY_PASSWORD`, or
- both `--handle` (`-u`) and `--password` (`-p`) directly on the command line.

## Behavior

- Verifies that the persistence layer exists (calls `setup.EnsurePersistenceReady`).
- Authenticates with the registered service using the handle/app password pair.
- Caches the resulting DID, tokens, and service URL inside `~/.skycli/.config.json`, encrypting the access and refresh tokens before writing.
- Marks the session as valid inside the database-backed session repository.

If authentication fails the command aborts without touching the existing session.

## Examples

```bash
# Explicit credentials
skycli login --handle @you.bsky.social --password app-password-1234

# Using a dotenv file
cat > ~/.config/.bsky.env <<'EOF'
BLUESKY_HANDLE=you.bsky.social
BLUESKY_PASSWORD=app-password-1234
EOF
skycli login --file ~/.config/.bsky.env
```

## Sample Output

```text
$ skycli login --handle @you.bsky.social --password app-password-1234
15:42:07 INFO  Authenticating with Bluesky handle=@you.bsky.social
✓ Successfully authenticated as @you.bsky.social
```

Timestamps and color styling come from `charmbracelet/log`; they appear on STDERR so you can pipe STDOUT for automation when needed.

## Troubleshooting

- “either --file or both --handle and --password are required” → provide credentials with one of the supported methods.
- “authentication succeeded but failed to save session” → disk permission issue preventing SkyCLI from updating `.config.json`; fix the permissions and retry.
- If your app password rotates, rerun `skycli login` with the new value to refresh the stored tokens.
