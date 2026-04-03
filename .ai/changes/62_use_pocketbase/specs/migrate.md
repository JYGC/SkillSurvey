# migrate — Spec (issue #62)

## Goals

- Provide a one-shot Go binary that reads all records from the legacy
  `SkillSurvey.db` SQLite database and writes them to `pocketbaseserver`.
- The binary is temporary: the operator deletes it and its directory after
  verifying the migration result.
- Integration tests must be written before implementation; no mocking.

## Module location

`migrate/` at the repository root (sibling of `pocketbaseserver/`).

Entry point: `migrate/cmd/migrate/main.go`.

## Dependencies

| Purpose | Package |
|---|---|
| Read legacy SQLite | `gorm.io/gorm` + `gorm.io/driver/sqlite` (same versions as `backend/`) |
| Write to PocketBase | `github.com/r--w/pocketbase` |
| Legacy entity models | Copy from `backend/internal/entities/` into `migrate/internal/legacyentities/` |

## Configuration

Read from a JSON file (`migrate.json`) located next to the executable.

```json
{
  "LegacyDbPath": "/path/to/SkillSurvey.db",
  "PocketBaseUrl": "http://...:8090",
  "ServiceAccountEmail": "...",
  "ServiceAccountPassword": "..."
}
```

Passwords are set by the operator; never hardcoded or committed.

## Migration steps

Run in this order to satisfy relation constraints:

1. **sites** — read `sites` table; create each record in PocketBase `sites`
   collection. Capture the old `ID → new PocketBase ID` mapping for use in step 4.

2. **skillTypes** — read `skill_types` table; create each record in
   `skillTypes`. Capture `ID → new ID` mapping for step 3.

3. **skillNames** — read `skill_names` table; look up `SkillTypeID` in the
   mapping from step 2; create each record in `skillNames` with the new
   `skillType` relation ID. Capture `ID → new ID` mapping for step 4 and 5.

4. **skillNameAliases** — read `skill_name_aliases` table; look up `SkillNameID`
   in the mapping from step 3; create each record in `skillNameAliases`.

5. **jobPosts** — read `job_posts` table; look up `SiteID` in the mapping from
   step 1; create each record in `jobPosts`. The `content` field must be a JSON
   object `{"title":"...","body":"..."}`. The `location` field must be a JSON
   object `{"city":"...","country":"...","suburb":"..."}`.

6. **monthlyCountReports** — read `monthly_count_reports` table; look up
   `SkillNameID` in the mapping from step 3; create each record in
   `monthlyCountReports`. The `identifier` field uses the **new** PocketBase
   `skillName` ID: `<newSkillNameId>_<YearMonth>`.

## Error handling

- If any step fails, log the error with the legacy record ID and continue
  migrating the remaining records (do not abort the entire run).
- Print a summary at the end: total records attempted vs. written per collection.
- Exit with a non-zero status if any record failed.

## Idempotency

The binary may be run more than once during testing. To avoid duplicate records
on re-runs, check for an existing record by a natural key before creating:

| Collection | Natural key |
|---|---|
| `sites` | `name` |
| `skillTypes` | `name` |
| `skillNames` | `(skillType, name)` |
| `skillNameAliases` | `(skillName, alias)` |
| `jobPosts` | `(site, jobSiteNumber)` |
| `monthlyCountReports` | `identifier` |

If a matching record already exists, skip creation and use its ID for the mapping.

## Build

- `migrate/Makefile` must provide a `build` target.
- No CGO: use `gorm.io/driver/sqlite` with `modernc.org/sqlite` instead of
  the cgo-based `go-sqlite3`. Update the GORM sqlite import accordingly.
- Must compile on OpenBSD: `GOOS=openbsd go build ./...` must succeed.

## Integration tests

Write tests in `migrate/` before implementing the migration logic:

- Given a legacy SQLite database with one record in each table, running
  the migration creates the corresponding records in a real PocketBase test
  instance (no mocking).
- Relations are correctly resolved: a `skillName` record in PocketBase
  references the newly created `skillType`, not the legacy numeric ID.
- Running the migration twice does not create duplicate records.
- A `monthlyCountReport` identifier uses the new PocketBase skill name ID,
  not the legacy integer ID.

## Out of scope

- Ongoing sync between the two databases.
- Rollback / undo of the migration (operator verifies and discards the binary).
- Migration of `users` or `userRoles` records (service accounts are created
  manually by the operator).
