# Design: Migrate to PocketBase (#62)

## Implementation order

Dependencies flow in this direction:

```
pocketbaseserver (migration + access rules)
    ↓
migrate/ (reads legacy SQLite, writes to pocketbaseserver)
    ↓
runtask/ (writes to pocketbaseserver at runtime)
    ↓
frontend (reads from pocketbaseserver via JS SDK)
```

Implement in that order so each phase can be tested against a running server with real data.

---

## 1. pocketbaseserver — add roles and access rules

### New file: `pocketbaseserver/migrations/1743552000_add_roles.go`

Package: `migrations`

#### `up` function steps

1. Create `roles` collection (`core.NewBaseCollection("roles")`):
   - Fields: `name` (TextField, required), `description` (TextField, required)
   - Unique index: `CREATE UNIQUE INDEX idx_roles_name ON roles (name ASC)`
   - ListRule/ViewRule: `@request.auth.id != ""`
   - CreateRule/UpdateRule/DeleteRule: `nil` (admin only)
   - Save the collection.

2. Insert three seed records into `roles` using `app.FindCollectionByNameOrId("roles")` then `core.NewRecord(rolesCollection)`:
   - `{name: "webscraper", description: "Write access to jobPosts"}`
   - `{name: "reporting", description: "Write access to monthlyCountReports"}`
   - `{name: "migration", description: "Write access to all collections except users, userRoles, and roles"}`

3. Create `userRoles` collection (`core.NewBaseCollection("userRoles")`):
   - Fields:
     - `user` (RelationField → `_pb_users_auth_`, MaxSelect=1, required)
     - `role` (RelationField → `roles` collection ID, MaxSelect=1, required)
   - Unique index: `CREATE UNIQUE INDEX idx_userRoles_user_role ON userRoles (user ASC, role ASC)`
   - ListRule/ViewRule: `@request.auth.id != ""`
   - CreateRule/UpdateRule/DeleteRule: `nil` (admin only)
   - Save the collection.

4. Apply write rules to existing collections. For each, call `app.FindCollectionByNameOrId(name)`, update the rule fields, then `app.Save(collection)`:

   | Collection | CreateRule / UpdateRule / DeleteRule |
   |---|---|
   | `jobPosts` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'webscraper' \|\| @collection.userRoles_via_user.role.name ?~ 'migration')` |
   | `monthlyCountReports` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'reporting' \|\| @collection.userRoles_via_user.role.name ?~ 'migration')` |
   | `skillTypes` | `@request.auth.id != "" && (@collection.userRoles_via_user.role.name ?~ 'migration' \|\| @request.auth.verified = true)` |
   | `skillNames` | same as `skillTypes` |
   | `skillNameAliases` | same as `skillTypes` |
   | `sites` | same as `skillTypes` |

#### `down` function steps

1. Revert write rules on `jobPosts`, `monthlyCountReports`, `skillTypes`, `skillNames`, `skillNameAliases`, `sites` (set CreateRule/UpdateRule/DeleteRule back to `nil`).
2. Delete seed records from `roles`.
3. Delete `userRoles` collection.
4. Delete `roles` collection.

### New models

**`pocketbaseserver/internal/models/role.go`**

```go
package models

type Role struct {
    Id          string
    Name        string
    Description string
}
```

**`pocketbaseserver/internal/models/userrole.go`**

```go
package models

type UserRole struct {
    Id   string
    User string // relation ID
    Role string // relation ID
}
```

### Integration tests

**`pocketbaseserver/migrations/1743552000_add_roles_test.go`** (write before implementing the migration):

Each test spins up a fresh in-process PocketBase app (same pattern as existing tests if any, otherwise use `tests.NewTestApp()` from `github.com/pocketbase/pocketbase/tests`), runs all migrations, then:

1. A user with no `userRoles` cannot write to `jobPosts` or `monthlyCountReports` — expect HTTP 403 from the collection API.
2. A user with the `webscraper` role can create a `jobPost` record — expect HTTP 200.
3. A user with the `reporting` role can create a `monthlyCountReport` record — expect HTTP 200.
4. A user with the `migration` role can create records in `sites`, `skillTypes`, `skillNames`, `skillNameAliases`, `jobPosts`, `monthlyCountReports` — expect HTTP 200 each.
5. A user with the `migration` role cannot create or modify `users`, `userRoles`, `roles` — expect HTTP 403.
6. Inserting a duplicate `(user, role)` pair into `userRoles` is rejected — expect an error.

---

## 2. migrate/ — one-shot legacy data migration

### Module layout

```
migrate/
  cmd/migrate/main.go
  internal/
    config/config.go
    legacyentities/
      entitybase.go          (copy from backend/internal/entities/entitybase.go)
      site.go
      skill.go               (SkillType, SkillName, SkillNameAlias)
      jobpost.go
      report.go
    migrator/
      migrator.go            (Migrator struct + Run method)
      sites.go
      skilltypes.go
      skillnames.go
      skillnamealiases.go
      jobposts.go
      monthlycountreports.go
  Makefile
  go.mod
```

### `migrate/go.mod`

Module: `keybook/migrate`

Dependencies:
- `gorm.io/gorm` (same version as `backend/`)
- `gorm.io/driver/sqlite` — **use `modernc.org/sqlite` tag, not `go-sqlite3`**: replace the import with the pure-Go driver (`gorm.io/driver/sqlite` supports `modernc` via build tag; alternatively use a fork that wraps `modernc.org/sqlite` directly)
- `github.com/r--w/pocketbase` — PocketBase Go client

> Note: `backend/go.mod` uses `gorm.io/driver/sqlite v1.1.6` with the CGO `go-sqlite3`. For `migrate/`, use `modernc.org/sqlite` as the GORM driver so the binary compiles without CGO on OpenBSD. Use `gorm.io/driver/sqlite` with a `modernc` replacement or the community wrapper `github.com/glebarez/sqlite` (which wraps `modernc.org/sqlite` and is API-compatible with `gorm.io/driver/sqlite`).

### `migrate/internal/config/config.go`

```go
type Config struct {
    LegacyDbPath         string
    PocketBaseUrl        string
    ServiceAccountEmail  string
    ServiceAccountPassword string
}
```

Load with `json.NewDecoder(file).Decode(&cfg)`. Config file path: `migrate.json` located next to the executable (use `os.Executable()` to find the directory).

### `migrate/internal/legacyentities/`

Copy the entity structs from `backend/internal/entities/` verbatim, updating the package name to `legacyentities`. No logic changes.

Files to copy:
- `entitybase.go` → keep `EntityBase` struct (GORM `Model` embed with `uint` primary key)
- `site.go` → `Site`
- `skill.go` → `SkillType`, `SkillName`, `SkillNameAlias`
- `jobpost.go` → `JobPost`
- `report.go` → `MonthlyCountReport`

### `migrate/internal/migrator/migrator.go`

```go
type Migrator struct {
    db  *gorm.DB
    pb  *pocketbase.Client
}

type Summary struct {
    Collection string
    Attempted  int
    Written    int
}

func New(db *gorm.DB, pb *pocketbase.Client) *Migrator

func (m *Migrator) Run() ([]Summary, error)
    // calls each step in order, accumulates summaries
```

`Run` calls steps in dependency order:
1. `migrateSites` → returns `map[uint]string` (legacyID → new PB ID)
2. `migrateSkillTypes` → returns `map[uint]string`
3. `migrateSkillNames(skillTypeIdMap)` → returns `map[uint]string`
4. `migrateSkillNameAliases(skillNameIdMap)`
5. `migrateJobPosts(siteIdMap)`
6. `migrateMonthlyCountReports(skillNameIdMap)`

### Per-collection migration pattern

Each step follows this pattern (shown for `sites`):

```
1. Read all records from legacy DB via GORM.
2. For each record:
   a. Check if it already exists in PocketBase by natural key (idempotency).
   b. If it exists, record its PB ID in the mapping and skip.
   c. If it does not exist, create it via the PocketBase client.
   d. On error, log the legacy record ID and continue; increment failed count.
3. Return the legacy-ID → new-PB-ID mapping.
```

#### Natural key filter per collection

| Collection | PocketBase filter string |
|---|---|
| `sites` | `name = "{name}"` |
| `skillTypes` | `name = "{name}"` |
| `skillNames` | `skillType = "{skillTypeId}" && name = "{name}"` |
| `skillNameAliases` | `skillName = "{skillNameId}" && alias = "{alias}"` |
| `jobPosts` | `site = "{siteId}" && jobSiteNumber = "{number}"` |
| `monthlyCountReports` | `identifier = "{identifier}"` |

Use `pb.List("collection", pocketbase.ParamsList{Filter: "..."})` to check existence before creating.

#### Field mappings

**jobPosts** — the legacy entity has flat fields; PocketBase stores JSON:
- `content`: `{"title": "<Title>", "body": "<Body>"}`
- `location`: `{"city": "<City>", "country": "<Country>", "suburb": "<Suburb>"}`

**monthlyCountReports** — `identifier` uses the **new** PocketBase skill name ID:
- `identifier`: `<newSkillNameId>_<YearMonth>`

### Error handling

Each step: if a record fails, `log.Printf("migrate %s id=%d: %v", collection, legacyID, err)` and continue. At the end, `Run` returns a `[]Summary` with per-collection counts. `main.go` prints the summary table and exits non-zero if any count shows failures.

### `migrate/cmd/migrate/main.go`

1. Load config from `migrate.json` next to the executable.
2. Open legacy SQLite with GORM (no CGO driver).
3. Authenticate with PocketBase: `pb.Authenticate(email, password)`.
4. `migrator.New(db, pb).Run()`.
5. Print summary. Exit 1 if any failures.

### `migrate/Makefile`

```makefile
OUTPUT_DIR=./build

mk_output_dir:
	mkdir -p ${OUTPUT_DIR}

build: mk_output_dir
	go build -o ${OUTPUT_DIR}/migrate ./cmd/migrate/main.go
```

### Integration tests

**`migrate/internal/migrator/migrator_test.go`** (write before implementing):

Test setup helper:
- Start a real PocketBase test instance in-process (`tests.NewTestApp()`), run migrations.
- Open an in-memory (or temp file) SQLite database with the legacy schema via GORM, seed with one record per table.

Tests:
1. After `Run()`, each PocketBase collection contains one record with correctly mapped fields.
2. Relations are resolved: `skillName.skillType` holds the new PocketBase ID (not the legacy integer).
3. `Run()` twice produces no duplicates — counts stay at one per collection.
4. A `monthlyCountReport` record's `identifier` contains the new PocketBase skill name ID, not the legacy integer.

---

## 3. runtask/ — replacement for backend tasks

### Module layout

```
runtask/
  cmd/runtask/main.go
  internal/
    config/config.go
    exception/
      errorlogging.go        (ported from backend/internal/exception/)
      errormessages.go
    dynamiccontentextractor/
      dynamiccontentextractor.go   (copied from backend/internal/dynamiccontentextractor/)
    siteadapters/
      siteadapterbase.go     (interface updated to return runtask job post type)
      seekadapter.go         (ported from backend/internal/siteadapters/)
      seekadapterconfig.go
      seekapigetparameters.go
      joraadapter.go
      joraadapterconfig.go
    pbclient/
      pbclient.go            (thin wrapper: auth + collection helpers)
    scrape/
      scrape.go              (scrape command logic)
    report/
      report.go              (report command logic)
    housekeeping/
      housekeeping.go        (cleanfs + sendlog)
  Makefile
  go.mod
```

### `runtask/go.mod`

Module: `keybook/runtask`

Dependencies:
- `github.com/chromedp/chromedp` (same version as `backend/`)
- `github.com/chromedp/cdproto` (same version as `backend/`)
- `github.com/r--w/pocketbase` — PocketBase Go client
- No CGO; no `gorm.io/driver/sqlite`

### `runtask/internal/config/config.go`

```go
type Config struct {
    PocketBaseUrl          string
    ServiceAccountEmail    string
    ServiceAccountPassword string
    SeekConfigFile         string
    JoraConfigFile         string
    ErrorLogFile           string
    SmtpHost               string
    SmtpPort               int
    EmailRecipient         string
}
```

Load from `runtask.json` next to the executable (via `os.Executable()`).

### `runtask/internal/exception/`

Copy `backend/internal/exception/errorlogging.go` and `errormessages.go` verbatim into `runtask/internal/exception/`. Update the package import path. Remove the `environment.AttachToExecutableDir` call — use the `ErrorLogFile` path from config instead (pass it as a parameter or initialise the logger with a path set from config at startup).

### `runtask/internal/siteadapters/`

Copy from `backend/internal/siteadapters/` verbatim. Update:
- Package import path from `github.com/JYGC/SkillSurvey/internal/...` to `keybook/runtask/internal/...`.
- `ISiteAdapter.RunSurvey()` still returns a slice of job post structs. Define a local `InboundJobPost` struct in `runtask/internal/siteadapters/` that mirrors the relevant fields (Title, Body, JobSiteNumber, PostedDate, City, Country, Suburb, SiteName) — removing the GORM entity dependency.

### `runtask/internal/dynamiccontentextractor/`

Copy from `backend/internal/dynamiccontentextractor/` verbatim. Update package import path only.

### `runtask/internal/pbclient/pbclient.go`

Thin wrapper around `github.com/r--w/pocketbase`:

```go
type Client struct {
    pb *pocketbase.Client
}

func New(url, email, password string) (*Client, error)
    // authenticates, returns error if auth fails

func (c *Client) GetSites() ([]Site, error)
func (c *Client) UpsertJobPost(post JobPost) error
    // filter by site+jobSiteNumber; skip if exists, create if not
func (c *Client) GetEnabledSkillNamesWithAliases() ([]SkillNameWithAliases, error)
func (c *Client) GetAllJobPosts() ([]JobPost, error)
func (c *Client) UpsertMonthlyCountReport(report MonthlyCountReport) error
    // filter by identifier; update if exists, create if not
```

Local struct definitions (no GORM, no PocketBase server dependency):
- `Site{Id, Name, Url string}`
- `JobPost{Id, JobSiteNumber, SiteId string; Content JobPostContent; Location JobPostLocation; PostedDate time.Time}`
- `JobPostContent{Title, Body string}`
- `JobPostLocation{City, Country, Suburb string}`
- `SkillNameWithAliases{Id, Name string; Aliases []string}`
- `MonthlyCountReport{Identifier, YearMonth string; YearMonthDate time.Time; Count int; SkillNameId string}`

### `runtask/internal/scrape/scrape.go`

```go
func Run(cfg config.Config, pb *pbclient.Client) error
```

Logic:
1. `pb.GetSites()` to get the site list.
2. For each site, instantiate the correct adapter (Seek or Jora) based on `site.Name` matching config file names.
3. `adapter.RunSurvey()` returns `[]InboundJobPost`.
4. For each result, call `pb.UpsertJobPost(...)` — builds `content` and `location` JSON structs.
5. Errors are written to `ErrorLogFile` via the exception logger; do not panic; continue processing remaining posts.

### `runtask/internal/report/report.go`

```go
func Run(cfg config.Config, pb *pbclient.Client) error
```

Logic:
1. `pb.GetEnabledSkillNamesWithAliases()` — fetches `skillNames` with filter `isEnabled = true`, expands `skillNameAliases_via_skillName`, builds `[]SkillNameWithAliases`.
2. `pb.GetAllJobPosts()` — fetches full list.
3. For each skill name, for each distinct `YearMonth` present in job posts, count how many job posts have `content.body` matching any alias using the word-boundary patterns ported from `backend/internal/database/jobposttablecall.go`:
   - `" alias "`, `",alias,"`, `".alias."`, `"\nalias\n"`, plus leading/trailing variants (`" alias,"`, `" alias."`, etc.).
   - Use `strings.Contains` (case-insensitive via `strings.ToLower`).
4. Build `identifier = <skillNameId>_<YYYY-MM>`.
5. Parse `YearMonthDate` with `time.Parse("2006-01", yearMonth)` then add day-01 to get first of month.
6. `pb.UpsertMonthlyCountReport(...)` for each result.

#### Word-boundary matching detail

Port this exact set of suffix/prefix combinations (matching the GORM LIKE query in `backend`):
```
" alias ", " alias,", " alias.", " alias\n",
",alias ", ",alias,", ",alias.", ",alias\n",
".alias ", ".alias,", ".alias.", ".alias\n",
"\nalias ", "\nalias,", "\nalias.", "\nalias\n"
```
Lower-case both the alias and the `content.body` before comparing.

### `runtask/internal/housekeeping/housekeeping.go`

```go
func CleanFS() error      // port cleanupFilesystem() from backend/cmd/housekeeping/main.go
func SendLog(cfg config.Config) error  // port sendLogToAdmin(); read ErrorLogFile, send via SMTP, truncate
```

`SendLog` uses `net/smtp` directly (same as the backend). SMTP config comes from `config.Config` fields (`SmtpHost`, `SmtpPort`, `EmailRecipient`).

### `runtask/cmd/runtask/main.go`

```go
func main() {
    // os.Args[1] selects the sub-command: "scrape", "report", "housekeeping cleanfs", "housekeeping sendlog"
    // Load config, authenticate pbclient, dispatch
}
```

Sub-commands:
- `scrape` → `scrape.Run(cfg, pb)`
- `report` → `report.Run(cfg, pb)`
- `housekeeping cleanfs` → `housekeeping.CleanFS()`
- `housekeeping sendlog` → `housekeeping.SendLog(cfg)`

All panics deferred to `exception.ReportErrorIfPanic(...)`.

### `runtask/Makefile`

```makefile
OUTPUT_DIR=./build

mk_output_dir:
	mkdir -p ${OUTPUT_DIR}

build: mk_output_dir
	go build -o ${OUTPUT_DIR}/runtask ./cmd/runtask/main.go

run_dev: build
	${OUTPUT_DIR}/runtask
```

### Integration tests

**Write tests before implementation.**

**`runtask/internal/scrape/scrape_test.go`**:
- Spin up a real PocketBase test instance with migrations applied (including `1743552000_add_roles.go`).
- Create a service account with `webscraper` role.
- Seed a `sites` record for "TestSite".
- Run a stub HTTP server (using `net/http/httptest`) that returns a known HTML page.
- Point the Seek or Jora adapter at the stub server.
- Call `scrape.Run(...)`.
- Assert `jobPosts` collection has the expected records.
- Run `scrape.Run(...)` a second time; assert record count did not increase.

**`runtask/internal/report/report_test.go`**:
- Spin up PocketBase test instance.
- Seed `skillNames`, `skillNameAliases`, and `jobPosts` with known content.
- Call `report.Run(...)`.
- Assert `monthlyCountReports` records match expected counts for each skill/month.

**`runtask/internal/housekeeping/housekeeping_test.go`**:
- `CleanFS`: create temp directories matching the Chromium patterns; call `CleanFS()`; assert they no longer exist.
- `SendLog`: write known content to a temp `error.log`; start an in-process SMTP test server using `github.com/emersion/go-smtp`; call `SendLog(...)`; assert one email received with the log content; assert `error.log` is now zero bytes.

---

## 4. frontend — migrate API calls

### Files to delete

```
src/views/public/SkillTypeAdd.vue
src/views/public/SkillTypeEdit.vue
src/views/public/SkillAdd.vue
src/views/public/SkillEdit.vue
src/views/public/SkillTypeList.vue
src/views/public/SkillList.vue
src/components/SkillTypeView.vue
src/components/SkillView.vue
```

`SkillTypeView.vue` is only used by `SkillTypeAdd.vue` and `SkillTypeEdit.vue`. `SkillView.vue` is only used by `SkillAdd.vue` and `SkillEdit.vue`. All skill and skill-type management (list, create, edit, delete) is handled via the PocketBase admin UI instead.

### Router (`src/main.ts`)

Remove route entries for `skill-add`, `skill-edit`, `skill-type-add`, `skill-type-edit`, `skill-list`, `skill-type-list`. Remove any navigation links that reference these routes.

### `src/schemas/skills.ts`

Replace entirely with:

```typescript
export interface SkillType {
  id: string;
  name: string;
  description: string;
  expand?: { skillNames_via_skillType?: SkillName[] };
}

export interface SkillName {
  id: string;
  skillType: string;
  name: string;
  isEnabled: boolean;
  expand?: {
    skillType?: SkillType;
    skillNameAliases_via_skillName?: SkillNameAlias[];
  };
}

export interface SkillNameAlias {
  id: string;
  skillName: string;
  alias: string;
  expand?: { skillName?: SkillName };
}
```

### `src/views/public/MonthlyCountReport.vue`

- Remove `const getMonthlyCountUrl` and the `fetch(...)` call.
- Import `getBackendClient`.
- Define a local interface:
  ```typescript
  interface MonthlyCountRecord {
    YearMonth: string;
    count: number;
    expand?: { skillName?: { name: string } };
  }
  ```
- Replace the fetch block with:
  ```typescript
  const allRecords = await getBackendClient()
    .collection('monthlyCountReports')
    .getFullList<MonthlyCountRecord>({
      expand: 'skillName',
      sort: 'yearMonthDate',
    });
  ```
- Filter to the most recent 12 distinct `YearMonth` values before building chart data:
  ```typescript
  const recentMonths = [...new Set(allRecords.map(r => r.YearMonth))]
    .slice(-12);
  const filtered = allRecords.filter(r => recentMonths.includes(r.YearMonth));
  ```
- Replace `createChartLabels()` with the 12 distinct `recentMonths` values (already sorted by `yearMonthDate`).
- Replace `createDataSet(data)` — group `filtered` by `record.expand?.skillName?.name`, build datasets using `record.YearMonth` and `record.count`.

### Non-functional

- Verify no remaining `http://localhost:3000` strings: `grep -r "localhost:3000" src/`.
- Run `npm run lint` in `frontend/` and fix all reported errors.
- No new npm dependencies.

---

## 5. Build verification

After all components are implemented, verify:

```sh
# pocketbaseserver
cd pocketbaseserver && make build

# migrate (no CGO)
cd migrate && GOOS=openbsd CGO_ENABLED=0 go build ./...

# runtask (no CGO)
cd runtask && GOOS=openbsd CGO_ENABLED=0 go build ./...

# frontend
cd frontend && npm run lint && npm run build
```

All four must succeed without errors.

---

## 6. Deployment sequence (operator steps, not automated)

1. Stop `backend/` services.
2. Build and start `pocketbaseserver` (`make build && make run_dev` or production equivalent).
3. PocketBase runs migrations automatically on startup — `roles`, `userRoles`, and access rules are applied.
4. Operator creates service accounts via the PocketBase admin UI:
   - One account for `runtask` with `webscraper` + `reporting` roles.
   - One account for `migrate` with the `migration` role.
5. Operator sets passwords in `runtask.json` and `migrate.json` next to the respective executables.
6. Run `./migrate` to copy legacy data; verify result in PocketBase admin UI; delete `migrate/` directory.
7. Deploy updated `frontend/` build.
8. Replace `backend/` cron jobs with equivalent `runtask scrape`, `runtask report`, `runtask housekeeping sendlog`, `runtask housekeeping cleanfs` invocations.
