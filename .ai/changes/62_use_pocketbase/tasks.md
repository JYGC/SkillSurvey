# Tasks: Migrate to PocketBase (#62)

Implementation order follows dependency flow:
`pocketbaseserver` → `migrate` → `runtask` → `frontend`

Tests must be written **before** implementation per rules.md. No database mocking.

---

## Phase 1: pocketbaseserver — roles and access rules

### Models
- [ ] Create `pocketbaseserver/internal/models/role.go` — package `models`; struct `Role { Id, Name, Description string }`
- [ ] Create `pocketbaseserver/internal/models/userrole.go` — package `models`; struct `UserRole { Id, User, Role string }` where `User` and `Role` are PocketBase relation IDs

### Integration tests (write first)
- [ ] Create `pocketbaseserver/migrations/1743552000_add_roles_test.go` — spin up `tests.NewTestApp()`, run all migrations, then assert:
  - [ ] A user with no `userRoles` row cannot POST to `jobPosts` or `monthlyCountReports` → expect HTTP 403
  - [ ] A user assigned the `webscraper` role can POST a valid `jobPost` record → expect HTTP 200
  - [ ] A user assigned the `reporting` role can POST a valid `monthlyCountReport` record → expect HTTP 200
  - [ ] A user assigned the `migration` role can POST to `sites`, `skillTypes`, `skillNames`, `skillNameAliases`, `jobPosts`, and `monthlyCountReports` → expect HTTP 200 for each
  - [ ] A user assigned the `migration` role cannot POST/PATCH/DELETE rows in `users`, `userRoles`, or `roles` → expect HTTP 403 for each
  - [ ] Attempting to insert a second row with the same `(user, role)` into `userRoles` returns a unique-constraint error

### Migration file: `pocketbaseserver/migrations/1743552000_add_roles.go`
- [ ] `up` step 1 — create `roles` collection via `core.NewBaseCollection("roles")`:
  - Fields: `name` (TextField, required), `description` (TextField, required)
  - Unique index: `CREATE UNIQUE INDEX idx_roles_name ON roles (name ASC)`
  - ListRule/ViewRule: `@request.auth.id != ""`; CreateRule/UpdateRule/DeleteRule: `nil`
  - Call `app.Save(rolesCollection)`
- [ ] `up` step 2 — insert seed records into `roles` using `app.FindCollectionByNameOrId("roles")` + `core.NewRecord(rolesCollection)`:
  - `{ name: "webscraper", description: "Write access to jobPosts" }`
  - `{ name: "reporting", description: "Write access to monthlyCountReports" }`
  - `{ name: "migration", description: "Write access to all collections except users, userRoles, and roles" }`
- [ ] `up` step 3 — create `userRoles` collection via `core.NewBaseCollection("userRoles")`:
  - Fields: `user` (RelationField → `_pb_users_auth_`, MaxSelect=1, required), `role` (RelationField → roles collection ID, MaxSelect=1, required)
  - Unique index: `CREATE UNIQUE INDEX idx_userRoles_user_role ON userRoles (user ASC, role ASC)`
  - ListRule/ViewRule: `@request.auth.id != ""`; CreateRule/UpdateRule/DeleteRule: `nil`
  - Call `app.Save(userRolesCollection)`
- [ ] `up` step 4 — apply access rules to existing collections via `app.FindCollectionByNameOrId(name)` + `app.Save(collection)`:
  - `monthlyCountReports`: ListRule/ViewRule = `""` (public); CreateRule/UpdateRule/DeleteRule = `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'reporting' || @collection.userRoles_via_user.role.name ?~ 'migration')`
  - `jobPosts`: CreateRule/UpdateRule/DeleteRule = `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'webscraper' || @collection.userRoles_via_user.role.name ?~ 'migration')`
  - `skillTypes`, `skillNames`, `skillNameAliases`, `sites`: CreateRule/UpdateRule/DeleteRule = `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'migration' || @request.auth.verified = true)`
- [ ] `down` step 1 — revert CreateRule/UpdateRule/DeleteRule to `nil` on all six collections; revert `monthlyCountReports` ListRule/ViewRule to `nil`
- [ ] `down` step 2 — delete the three seed records from `roles`
- [ ] `down` step 3 — delete `userRoles` collection
- [ ] `down` step 4 — delete `roles` collection

### Verification
- [ ] `cd pocketbaseserver && make build` — binary compiles without errors

---

## Phase 2: migrate — one-shot legacy data migration

### Module scaffolding
- [ ] Create `migrate/go.mod` — module `keybook/migrate`; require `gorm.io/gorm`, `gorm.io/driver/sqlite` (same versions as `backend/`, CGO `go-sqlite3` driver), `github.com/r--w/pocketbase`
- [ ] Create `migrate/Makefile`:
  ```makefile
  OUTPUT_DIR=./build
  mk_output_dir:
      mkdir -p ${OUTPUT_DIR}
  build: mk_output_dir
      go build -o ${OUTPUT_DIR}/migrate ./cmd/migrate/main.go
  ```

### Config
- [ ] Create `migrate/internal/config/config.go` — struct `Config { LegacyDbPath, PocketBaseUrl, ServiceAccountEmail, ServiceAccountPassword string }`; load with `json.NewDecoder(file).Decode(&cfg)` from `migrate.json` adjacent to the executable (use `os.Executable()` to resolve the directory)

### Legacy entities — copy from `backend/internal/entities/`, change package to `legacyentities`, no other changes
- [ ] Create `migrate/internal/legacyentities/entitybase.go` — `EntityBase` struct (GORM `Model` embed with `uint` primary key)
- [ ] Create `migrate/internal/legacyentities/site.go` — `Site` struct
- [ ] Create `migrate/internal/legacyentities/skill.go` — `SkillType`, `SkillName`, `SkillNameAlias` structs
- [ ] Create `migrate/internal/legacyentities/jobpost.go` — `JobPost` struct (flat fields: Title, Body, JobSiteNumber, PostedDate, City, Country, Suburb, SiteID)
- [ ] Create `migrate/internal/legacyentities/report.go` — `MonthlyCountReport` struct

### Integration tests (write first)
- [ ] Create `migrate/internal/migrator/migrator_test.go`:
  - [ ] Helper: open in-memory/temp SQLite, create legacy schema via GORM AutoMigrate, seed one record per table with known field values; spin up `tests.NewTestApp()` with all pocketbaseserver migrations applied
  - [ ] Test: after `Run()`, each of the six PocketBase collections contains exactly one record; field values match the seeded legacy values
  - [ ] Test: `skillName` record in PocketBase has `skillType` set to the new PocketBase `skillTypes` record ID, not the legacy integer
  - [ ] Test: calling `Run()` a second time leaves record counts unchanged (idempotent)
  - [ ] Test: `monthlyCountReport.identifier` is `<newSkillNameId>_<YearMonth>` using the new PocketBase ID, not the legacy integer

### Migrator
- [ ] Create `migrate/internal/migrator/migrator.go`:
  - `Migrator struct { db *gorm.DB; pb *pocketbase.Client }`
  - `Summary struct { Collection string; Attempted, Written int }`
  - `func New(db *gorm.DB, pb *pocketbase.Client) *Migrator`
  - `func (m *Migrator) Run() ([]Summary, error)` — calls each step in dependency order, accumulates summaries
- [ ] Create `migrate/internal/migrator/sites.go` — read all `Site` rows from GORM; for each, check PocketBase via `pb.List("sites", {Filter: 'name = "<name>"'})`; skip and record ID if found, create otherwise; return `map[uint]string`
- [ ] Create `migrate/internal/migrator/skilltypes.go` — same pattern; natural key filter: `name = "<name>"`; return `map[uint]string`
- [ ] Create `migrate/internal/migrator/skillnames.go` — resolve `SkillTypeID` through `skillTypeIdMap`; natural key filter: `skillType = "<id>" && name = "<name>"`; return `map[uint]string`
- [ ] Create `migrate/internal/migrator/skillnamealiases.go` — resolve `SkillNameID` through `skillNameIdMap`; natural key filter: `skillName = "<id>" && alias = "<alias>"`
- [ ] Create `migrate/internal/migrator/jobposts.go` — resolve `SiteID` through `siteIdMap`; natural key filter: `site = "<id>" && jobSiteNumber = "<number>"`; encode `content` as `{"title":"...","body":"..."}` JSON and `location` as `{"city":"...","country":"...","suburb":"..."}` JSON before POSTing
- [ ] Create `migrate/internal/migrator/monthlycountreports.go` — resolve `SkillNameID` through `skillNameIdMap`; build `identifier` as `<newSkillNameId>_<YearMonth>` (not legacy integer); natural key filter: `identifier = "<value>"`
- [ ] Error handling in all steps: on failure, `log.Printf("migrate %s id=%d: %v", collection, legacyID, err)` and continue; increment `Summary.Attempted` regardless, increment `Summary.Written` only on success

### Entry point
- [ ] Create `migrate/cmd/migrate/main.go`:
  1. Load `Config` from `migrate.json` next to the executable
  2. Open legacy SQLite with `gorm.Open(sqlite.Open(cfg.LegacyDbPath), &gorm.Config{})` (no AutoMigrate — read-only)
  3. Create `pocketbase.NewClient(cfg.PocketBaseUrl)` and call `pb.Authenticate(email, password)`
  4. Call `migrator.New(db, pb).Run()`
  5. Print summary table (collection, attempted, written) to stdout
  6. Exit 1 if any collection has `Attempted != Written`

### Verification
- [ ] `cd migrate && GOOS=openbsd go build ./...` — compiles cleanly (CGO `go-sqlite3` is acceptable on OpenBSD; production host has the C toolchain)

---

## Phase 3: runtask — replacement for backend tasks

### Module scaffolding
- [ ] Create `runtask/go.mod` — module `keybook/runtask`; require `github.com/chromedp/chromedp`, `github.com/chromedp/cdproto` (same versions as `backend/`), `github.com/r--w/pocketbase`; no `gorm.io/driver/sqlite`, no CGO
- [ ] Create `runtask/Makefile`:
  ```makefile
  OUTPUT_DIR=./build
  mk_output_dir:
      mkdir -p ${OUTPUT_DIR}
  build: mk_output_dir
      go build -o ${OUTPUT_DIR}/runtask ./cmd/runtask/main.go
  run_dev: build
      ${OUTPUT_DIR}/runtask
  ```

### Config
- [ ] Create `runtask/internal/config/config.go` — struct `Config { PocketBaseUrl, ServiceAccountEmail, ServiceAccountPassword, SeekConfigFile, JoraConfigFile, ErrorLogFile, SmtpHost string; SmtpPort int; EmailRecipient string }`; load from `runtask.json` adjacent to the executable (use `os.Executable()`)

### Exception handling
- [ ] Create `runtask/internal/exception/errorlogging.go` — port from `backend/internal/exception/errorlogging.go`; replace `environment.AttachToExecutableDir` with a plain file path parameter; initialize the logger with `ErrorLogFile` from config at startup
- [ ] Create `runtask/internal/exception/errormessages.go` — port verbatim, update package import path from `github.com/JYGC/SkillSurvey/internal/...` to `keybook/runtask/internal/...`

### Dynamic content extractor
- [ ] Create `runtask/internal/dynamiccontentextractor/dynamiccontentextractor.go` — copy `backend/internal/dynamiccontentextractor/` verbatim; update package import path only

### Site adapters
- [ ] Define `InboundJobPost` struct in `runtask/internal/siteadapters/` — fields: `Title, Body, JobSiteNumber string; PostedDate time.Time; City, Country, Suburb, SiteName string` (no GORM dependency)
- [ ] Create `runtask/internal/siteadapters/siteadapterbase.go` — `ISiteAdapter` interface; `RunSurvey()` returns `([]InboundJobPost, error)`
- [ ] Create `runtask/internal/siteadapters/seekadapter.go` — copy from `backend/internal/siteadapters/seekadapter.go`; update import paths; update `RunSurvey()` signature to return `[]InboundJobPost`
- [ ] Create `runtask/internal/siteadapters/seekadapterconfig.go` — copy verbatim, update import path
- [ ] Create `runtask/internal/siteadapters/seekapigetparameters.go` — copy verbatim, update import path
- [ ] Create `runtask/internal/siteadapters/joraadapter.go` — copy from `backend/internal/siteadapters/joraadapter.go`; update import paths; update `RunSurvey()` signature
- [ ] Create `runtask/internal/siteadapters/joraadapterconfig.go` — copy verbatim, update import path

### PocketBase client
- [ ] Create `runtask/internal/pbclient/pbclient.go`:
  - Local structs: `Site { Id, Name, Url string }`, `JobPostContent { Title, Body string }`, `JobPostLocation { City, Country, Suburb string }`, `JobPost { Id, JobSiteNumber, SiteId string; Content JobPostContent; Location JobPostLocation; PostedDate time.Time }`, `SkillNameWithAliases { Id, Name string; Aliases []string }`, `MonthlyCountReport { Identifier, YearMonth string; YearMonthDate time.Time; Count int; SkillNameId string }`
  - `func New(url, email, password string) (*Client, error)` — creates `pocketbase.NewClient(url)`, calls `Authenticate`, returns error if auth fails
  - `func (c *Client) GetSites() ([]Site, error)` — `getFullList("sites")`
  - `func (c *Client) UpsertJobPost(post JobPost) error` — filter `site = "<id>" && jobSiteNumber = "<num>"`; if list returns ≥1 result, return nil (skip); otherwise create
  - `func (c *Client) GetEnabledSkillNamesWithAliases() ([]SkillNameWithAliases, error)` — `getFullList("skillNames", filter: "isEnabled = true", expand: "skillNameAliases_via_skillName")`; build `Aliases` slice from expanded records
  - `func (c *Client) GetAllJobPosts() ([]JobPost, error)` — `getFullList("jobPosts")`; unmarshal `content` and `location` JSON fields
  - `func (c *Client) UpsertMonthlyCountReport(report MonthlyCountReport) error` — filter `identifier = "<value>"`; if found, update (PATCH); if not, create (POST)

### Integration tests (write first)
- [ ] Create `runtask/internal/scrape/scrape_test.go`:
  - [ ] Setup: `tests.NewTestApp()` with all pocketbaseserver migrations; create service account with `webscraper` role; seed a `sites` record for "TestSite"; start `httptest.NewServer(...)` returning known HTML matching Seek/Jora response structure; point adapter config at stub server URL
  - [ ] Test: call `scrape.Run(cfg, pb)` → assert `jobPosts` collection contains expected records (count, field values)
  - [ ] Test: call `scrape.Run(cfg, pb)` a second time → assert `jobPosts` count is unchanged (idempotent upsert)
- [ ] Create `runtask/internal/report/report_test.go`:
  - [ ] Setup: `tests.NewTestApp()`; seed `skillTypes`, `skillNames`, `skillNameAliases`, and `jobPosts` with content bodies known to contain / not contain specific alias patterns across multiple months
  - [ ] Test: call `report.Run(cfg, pb)` → assert `monthlyCountReports` records match expected per-skill-per-month counts
- [ ] Create `runtask/internal/housekeeping/housekeeping_test.go`:
  - [ ] Test `CleanFS`: create temp directory tree matching Chromium temp file patterns; call `CleanFS()`; assert matching directories/files no longer exist
  - [ ] Test `SendLog`: write known string to a temp `error.log`; start in-process SMTP server using `github.com/emersion/go-smtp`; call `SendLog(cfg)`; assert exactly one message received at `EmailRecipient` whose body contains the log content; assert `error.log` is now zero bytes

### Scrape command
- [ ] Create `runtask/internal/scrape/scrape.go` — `func Run(cfg config.Config, pb *pbclient.Client) error`:
  1. `pb.GetSites()` → site list
  2. For each site, match `site.Name` against `cfg.SeekConfigFile` / `cfg.JoraConfigFile` to select adapter; instantiate via config file
  3. `adapter.RunSurvey()` → `[]InboundJobPost`
  4. For each post, call `pb.UpsertJobPost(...)` — build `JobPostContent{Title, Body}` and `JobPostLocation{City, Country, Suburb}` from `InboundJobPost`
  5. On error, write to `ErrorLogFile` via exception logger; continue remaining posts

### Report command
- [ ] Create `runtask/internal/report/report.go` — `func Run(cfg config.Config, pb *pbclient.Client) error`:
  1. `pb.GetEnabledSkillNamesWithAliases()` → skill list with aliases
  2. `pb.GetAllJobPosts()` → all job posts
  3. For each `(skillName, yearMonth)` pair, count posts whose lowercased `content.Body` contains any lowercased alias wrapped in any of the 16 word-boundary patterns: `{ , . \n}` × alias × `{ , . \n}` (port from `backend/internal/database/jobposttablecall.go`)
  4. Build `identifier = <skillNameId>_<YYYY-MM>`; parse `yearMonthDate` with `time.Parse("2006-01", yearMonth)` (first of month)
  5. `pb.UpsertMonthlyCountReport(...)` for each non-zero count

### Housekeeping command
- [ ] Create `runtask/internal/housekeeping/housekeeping.go`:
  - `func CleanFS() error` — port `cleanupFilesystem()` from `backend/cmd/housekeeping/main.go`; remove Chromium temp directories matching the same glob patterns
  - `func SendLog(cfg config.Config) error` — read full contents of `cfg.ErrorLogFile`; compose MIME email with `net/smtp`; send to `cfg.EmailRecipient` via `cfg.SmtpHost:cfg.SmtpPort`; truncate `ErrorLogFile` to zero bytes after successful send

### Entry point
- [ ] Create `runtask/cmd/runtask/main.go`:
  - Defer `exception.ReportErrorIfPanic(...)` at top of `main`
  - Load config from `runtask.json` next to executable
  - Authenticate `pbclient.New(url, email, password)` (required for `scrape` and `report`; skip for housekeeping commands)
  - Dispatch on `os.Args[1]` (and `os.Args[2]` for `housekeeping`):
    - `"scrape"` → `scrape.Run(cfg, pb)`
    - `"report"` → `report.Run(cfg, pb)`
    - `"housekeeping cleanfs"` → `housekeeping.CleanFS()`
    - `"housekeeping sendlog"` → `housekeeping.SendLog(cfg)`
    - unknown → print usage and exit 1

### Verification
- [ ] `cd runtask && GOOS=openbsd CGO_ENABLED=0 go build ./...` — compiles cleanly (no CGO; connects to system Chrome already on OpenBSD host)

---

## Phase 4: frontend — migrate API calls

### Delete obsolete files
- [ ] Delete `src/views/public/SkillTypeAdd.vue` — skill/skill-type CRUD is now handled via PocketBase admin UI
- [ ] Delete `src/views/public/SkillTypeEdit.vue`
- [ ] Delete `src/views/public/SkillAdd.vue`
- [ ] Delete `src/views/public/SkillEdit.vue`
- [ ] Delete `src/views/public/SkillTypeList.vue`
- [ ] Delete `src/views/public/SkillList.vue`
- [ ] Delete `src/components/SkillTypeView.vue` — only used by `SkillTypeAdd.vue` / `SkillTypeEdit.vue`
- [ ] Delete `src/components/SkillView.vue` — only used by `SkillAdd.vue` / `SkillEdit.vue`

### Router cleanup (`src/main.ts`)
- [ ] Remove `import` statements for all eight deleted files
- [ ] Remove route definitions for `skill-add`, `skill-edit`, `skill-type-add`, `skill-type-edit`, `skill-list`, `skill-type-list`
- [ ] Remove any `<router-link>` or `$router.push` navigation references to those routes

### Schema update (`src/schemas/skills.ts`)
- [ ] Replace file contents entirely with PocketBase-shaped interfaces per spec:
  - `SkillType { id: string; name: string; description: string; expand?: { skillNames_via_skillType?: SkillName[] } }`
  - `SkillName { id: string; skillType: string; name: string; isEnabled: boolean; expand?: { skillType?: SkillType; skillNameAliases_via_skillName?: SkillNameAlias[] } }`
  - `SkillNameAlias { id: string; skillName: string; alias: string; expand?: { skillName?: SkillName } }`

### `src/views/public/MonthlyCountReport.vue`
- [ ] Remove `const getMonthlyCountUrl` constant and the `fetch(...)` call that uses it
- [ ] Add `import { getBackendClient } from ...` (existing PocketBase JS SDK helper)
- [ ] Define local interface `MonthlyCountRecord { YearMonth: string; count: number; expand?: { skillName?: { name: string } } }`
- [ ] Replace fetch block with:
  ```typescript
  const allRecords = await getBackendClient()
    .collection('monthlyCountReports')
    .getFullList<MonthlyCountRecord>({ expand: 'skillName', sort: 'yearMonthDate' });
  ```
- [ ] Add post-fetch filter for most recent 12 months:
  ```typescript
  const recentMonths = [...new Set(allRecords.map(r => r.YearMonth))].slice(-12);
  const filtered = allRecords.filter(r => recentMonths.includes(r.YearMonth));
  ```
- [ ] Update chart label generation to use `recentMonths` array (already sorted by `yearMonthDate`)
- [ ] Update dataset construction to group `filtered` by `record.expand?.skillName?.name`; use `record.YearMonth` as x-axis key and `record.count` as value

### Non-functional checks
- [ ] Run `grep -r "localhost:3000" src/` — confirm zero matches
- [ ] `cd frontend && npm run lint` — passes with no errors
- [ ] Confirm `package.json` and `package-lock.json` have no new dependencies

### Verification
- [ ] `cd frontend && npm run build` — compiles without errors

---

## Phase 5: final build verification

- [ ] `cd pocketbaseserver && make build`
- [ ] `cd migrate && GOOS=openbsd go build ./...`
- [ ] `cd runtask && GOOS=openbsd CGO_ENABLED=0 go build ./...`
- [ ] `cd frontend && npm run lint && npm run build`
