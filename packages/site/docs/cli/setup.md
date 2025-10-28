---
sidebar_position: 1
title: Setup
---

# setup

Initialize the persistence layer SkyCLI depends on: the config directory, encrypted session file, and SQLite cache. Every other command expects this groundwork to be in place.

```bash
skycli setup
```

## What It Does

- Resolves the platform-specific config dir (`~/.skycli` on macOS/Linux, `%APPDATA%\skycli` on Windows).
- Creates the directory with `0700` permissions if missing.
- Creates (or verifies) the cache database at `cache.db`.
- Runs SQLite migrations via `store.RunMigrations`, bringing the schema up to the latest version.
- Prints status messages so you can tell whether actions were skipped or executed.

The command is idempotent—you can run it as many times as you like to confirm state. On an up-to-date database it exits early without work.

## Sample Output

```text
$ skycli setup
Setup: Initializing persistence layer

ℹ Config directory: /Users/alex/.skycli
ℹ Database path: /Users/alex/.skycli/cache.db

✓ Config directory created
ℹ Database does not exist, will be created

ℹ Running migrations...

✓ Setup complete!
ℹ Database version: v3
ℹ Migrations applied: 3
```

## Failure Modes

- Missing permissions while creating directories or files will stop the run; fix the filesystem ownership or rerun with elevated privileges.
- If migrations fail, the log includes the specific SQL migration error. You can re-run once the underlying issue is corrected.

If you see “persistence layer not ready” from other commands, re-run `skycli setup` to repair the state.
