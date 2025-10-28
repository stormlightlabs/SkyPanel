---
sidebar_position: 3
title: Status
---

# status

Summarize the active SkyCLI session so you can confirm which account and service endpoint are in use.

```bash
skycli status
```

## Behavior

- Ensures the persistence layer is initialized.
- Reads the stored session from the registry-backed repository.
- If no valid session exists, prints a friendly reminder to run `skycli login`.
- Otherwise emits a short table with the handle, service URL, and an “Authenticated” confirmation.

## Sample Output

```text
$ skycli status
Session Status
ℹ Handle: @you.bsky.social
ℹ Service: https://bsky.social
✓ Authenticated
```

If no session is cached you will instead see:

```text
ℹ Not authenticated. Run 'skycli login' to authenticate.
```

## Notes

- Status reads from local state only; it does not make network calls.
- Use it in scripts to gate commands that require authentication (`skycli status >/dev/null`).
