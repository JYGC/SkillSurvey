# migrate — Codebase Index

## Purpose

One-shot data migration tool. Reads from the legacy SQLite database (`SkillSurvey.db`) and writes records into PocketBase. Intended to run once when bootstrapping a fresh PocketBase instance. All migration logic is idempotent — re-running is safe.

## Tech stack

- Go 1.24.1
- GORM + `gorm.io/driver/sqlite` (reads legacy SQLite)
- `github.com/r--w/pocketbase` HTTP client (writes to PocketBase)
- Target OS: OpenBSD

## Directory map

```
migrate/
├── cmd/migrate/
│   └── main.go              # Entry point; prints per-collection summary
├── internal/
│   ├── config/
│   │   └── config.go        # Loads migrate.json (legacy DB path, PocketBase URL + creds)
│   ├── legacyentities/      # GORM models for legacy SQLite schema
│   │   ├── entitybase.go    # Shared EntityBase (ID, CreatedAt, UpdatedAt)
│   │   ├── site.go          # Site { Name, Url }
│   │   ├── skill.go         # SkillType, SkillName, SkillNameAlias
│   │   ├── jobpost.go       # JobPost with location fields
│   │   └── report.go        # MonthlyCountReport { SkillNameID, YearMonth, Count }
│   └── migrator/
│       ├── migrator.go          # Orchestrator: runs steps in dependency order, returns []Summary
│       ├── sites.go             # migrateSites()
│       ├── skilltypes.go        # migrateSkillTypes()
│       ├── skillnames.go        # migrateSkillNames()
│       ├── skillnamealiases.go  # migrateSkillNameAliases()
│       ├── jobposts.go          # migrateJobPosts()
│       ├── monthlycountreports.go  # migrateMonthlyCountReports()
│       └── migrator_test.go
├── Makefile
├── go.mod                   # module: keybook/migrate
└── go.sum
```

## Config (migrate.json)

```json
{
  "LegacyDbPath":           "/path/to/SkillSurvey.db",
  "PocketBaseUrl":          "http://192.168.8.147:8090",
  "ServiceAccountEmail":    "runtask@skillsurvey.com",
  "ServiceAccountPassword": "<password>"
}
```

The service account must have the `migration` role in PocketBase.

## Migration steps (in dependency order)

1. **sites** — deduplicated by `name`
2. **skillTypes** — deduplicated by `name`; empty description falls back to skill type name
3. **skillNames** — deduplicated by `skillType (PB ID) + name`; preserves old→new ID mapping
4. **skillNameAliases** — deduplicated by `skillName (PB ID) + alias`
5. **jobPosts** — deduplicated by `jobSiteNumber + site (PB ID)`
6. **monthlyCountReports** — identifier = `<new skillNamePBID>_<YearMonth>`; deduplicated by identifier

Each step builds an `oldID → PocketBase ID` map used by dependent steps to preserve relations.

## Output

```
sites:                attempted=2  written=2
skillTypes:           attempted=22 written=22
skillNames:           attempted=94 written=94
skillNameAliases:     attempted=107 written=107
jobPosts:             attempted=N  written=N
monthlyCountReports:  attempted=N  written=N
```

`written` counts both newly created records and records that already existed (idempotent skips still increment `written`).

## Tests

Integration tests in `migrator_test.go` start a real PocketBase server with all `pocketbaseserver/migrations` applied (including skill data seed). All tests are idempotent-aware:

- `TestMigratorRunCreatesAllRecords` — verifies every legacy record lands in PocketBase
- `TestMigratorSkillNameHasNewSkillTypeID` — verifies relations are rewired to PB IDs (filters by `name && skillType` since the seed migration also contains a "Go" skill)
- `TestMigratorIsIdempotent` — runs `m.Run()` twice and asserts no new records appear on the second run
- `TestMigratorMonthlyCountReportIdentifierUsesNewPBID` — verifies identifier uses the new PB ID, not the legacy integer
