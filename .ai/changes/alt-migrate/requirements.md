# Requirements: alt-migrate

## Background

The existing `migrate` tool populates PocketBase via HTTP API calls. For large collections — especially `jobPosts` — this is too slow. `alt-migrate` is a temporary command embedded in `pocketbaseserver` that reads the legacy SQLite backup and writes records directly through PocketBase's internal Go API, bypassing HTTP entirely.

---

## Command invocation

### Startup
WHEN `pocketbaseserver alt-migrate --db <path>` is run THE SYSTEM SHALL start PocketBase in non-serving mode, run the migration, print a summary, and exit.
WHEN the `--db` flag is omitted THE SYSTEM SHALL exit with a non-zero status and print usage instructions.
WHEN the file specified by `--db` does not exist or cannot be opened THE SYSTEM SHALL exit with a non-zero status and a descriptive error message before processing any records.

---

## Migration scope

### Collections
WHEN alt-migrate runs THE SYSTEM SHALL migrate `jobPosts` from the legacy SQLite database into PocketBase.
WHEN alt-migrate runs THE SYSTEM SHALL NOT migrate sites, skillTypes, skillNames, skillNameAliases, or monthlyCountReports — these are handled by the existing `migrate` tool, which is fast enough for those smaller collections.

### Write method
WHEN writing a jobPost THE SYSTEM SHALL use PocketBase's internal Go API (`app.Save()`) — not HTTP — so that PocketBase's schema validation and indexing are respected without the overhead of an HTTP round-trip.

---

## Site resolution

WHEN migrating a jobPost THE SYSTEM SHALL look up the corresponding PocketBase site record by matching the legacy site name.
WHEN no matching site is found in PocketBase THE SYSTEM SHALL skip that jobPost and log a warning including the legacy record ID and site name.

---

## Idempotency

WHEN a jobPost with the same `jobSiteNumber` and `site` (PocketBase ID) already exists THE SYSTEM SHALL skip it without error.
WHEN alt-migrate is run multiple times against the same legacy database THE SYSTEM SHALL produce the same final PocketBase state with no duplicate records.

---

## Output

WHEN migration completes THE SYSTEM SHALL print a one-line summary per collection migrated:

```
jobPosts:   attempted=N  written=N  skipped=N  failed=N
```

---

## Error handling

WHEN a write error occurs for an individual jobPost THE SYSTEM SHALL log the error with the legacy record ID, increment the failed counter, and continue processing remaining records.
WHEN migration finishes with one or more failed records THE SYSTEM SHALL exit with a non-zero status.
WHEN migration finishes with zero failures THE SYSTEM SHALL exit with status 0.

---

## Temporary status

WHEN the migration of all jobPost data is confirmed complete THE SYSTEM SHALL be removed from pocketbaseserver — it is not intended as a permanent feature.
