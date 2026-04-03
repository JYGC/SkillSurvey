# backend — Codebase Index

> **Status: Legacy.** Being replaced by `pocketbaseserver`. The frontend still calls the REST API at `:3000` for skill/report data.

## Purpose

Four Go binaries that crawl job listings, compute monthly skill-demand reports, serve a REST API, and perform maintenance.

```
survey ──► db ◄── reports ──► db
                              ▲
                     results (REST :3000)
                     housekeeping (cron)
```

## Tech stack

- Go 1.24, standard `net/http`
- GORM + SQLite (`SkillSurvey.db`, created next to executable)
- Chromedp (headless Chrome) for dynamic-page scraping
- Config via JSON files (no env vars for runtime config)
- Target OS: OpenBSD

## Directory map

```
backend/
├── cmd/
│   ├── survey/main.go          # Scraper: crawls Seek + Jora, inserts JobPosts
│   ├── reports/main.go         # Report generator: monthly skill counts
│   ├── results/main.go         # REST API server (port 3000)
│   └── housekeeping/main.go    # Maintenance: temp-file cleanup, email log
├── internal/
│   ├── config/                 # JSON config loader (configbase + mainconfig)
│   ├── database/               # GORM adapters for every table
│   │   ├── databaseadapter.go  # Singleton DbAdapter, opens SQLite DB
│   │   ├── jobposttablecall.go # BulkUpdateOrInsert, GetMonthlyCountBySkill
│   │   ├── skilltypetablecall.go
│   │   ├── skillnametablecall.go   # GetAllWithTypeAndAliases, SaveOneWithTypeAndAliases
│   │   └── monthlycountreporttablecall.go
│   ├── entities/               # GORM models
│   │   ├── jobpost.go          # JobPost, InboundJobPost
│   │   ├── skill.go            # SkillType, SkillName, SkillNameAlias
│   │   ├── site.go             # Site
│   │   └── report.go           # MonthlyCountReport
│   ├── handlers/               # HTTP handlers
│   │   ├── commonresponse.go   # JSON response helper
│   │   ├── report.go           # GET /report/getmonthlycount
│   │   ├── skill.go            # /skill/* CRUD
│   │   └── skilltype.go        # /skilltype/* CRUD
│   ├── siteadapters/           # Scraper implementations
│   │   ├── joraadapter.go      # Chromedp browser automation for au.jora.com
│   │   ├── seekadapter.go      # Seek internal API + Chromedp for detail pages
│   │   └── siteadapterbase.go  # ISiteAdapter interface
│   ├── dynamiccontentextractor/ # Chromedp wrapper (user-agent spoofing, wait)
│   ├── getapiscraper/          # Generic paginated GET-API scraper
│   ├── exception/              # Error logging (panic recovery + stack trace → error.log)
│   ├── environment/            # Executable directory resolution
│   ├── readonlysettings/       # App-data folder (Windows / Linux paths)
│   └── dataschemas/            # DTOs (AliasWithSkillName)
├── config.json                 # { "ServerPort": "3000" }
├── jora.json                   # Jora scraper config (selectors, pages, rate-limits)
├── seek.json                   # Seek scraper config (API params, selectors)
└── Makefile                    # build_survey | build_reports | build_results | build_housekeeping | deploy
```

## Data models (GORM)

| Entity | Key fields |
|---|---|
| `Site` | `Name` (seek/jora domain) |
| `JobPost` | `SiteID`, `JobSiteNumber` (unique/site), `Title`, `Body`, `PostedDate`, `City`, `Country`, `Suburb` |
| `SkillType` | `Name`, `Description` |
| `SkillName` | `SkillTypeID`, `Name`, `IsEnabled` |
| `SkillNameAlias` | `SkillNameID`, `Alias` |
| `MonthlyCountReport` | `Identifier` (SkillID+YearMonth), `SkillNameID`, `YearMonth` (YYYY-MM), `Count` |

## REST API (port 3000)

All responses are JSON. CORS enabled for all origins/methods.

| Method | Path | Handler |
|---|---|---|
| GET | `/report/getmonthlycount` | Monthly skill counts |
| GET | `/skilltype/getall` | All skill types with nested skills |
| GET | `/skilltype/getbyid?skilltypeid=` | Single skill type |
| GET | `/skilltype/getallidandname` | ID → name map |
| POST | `/skilltype/add` | Create skill type |
| POST | `/skilltype/save` | Update skill type |
| POST | `/skilltype/delete` | Delete skill type (fails if has skills) |
| GET | `/skill/getall` | All skills with type and aliases |
| GET | `/skill/getbyid?skillid=` | Single skill |
| POST | `/skill/add` | Create skill + aliases |
| POST | `/skill/save` | Update skill + aliases |
| POST | `/skill/delete` | Delete skill |

## Skill matching logic

Job post `body` searched for each alias using case-insensitive SQLite LIKE with boundary patterns: ` alias `, `,alias`, `.alias`, `\nalias`.

## Config files needed at runtime

| File | Used by |
|---|---|
| `config.json` | `results` — server port |
| `jora.json` | `survey` — Jora scraper settings |
| `seek.json` | `survey` — Seek scraper settings |
| `mailadmin.json` | `housekeeping` — email credentials |

## Build & run

```bash
make build_survey        # → ./build/survey
make build_reports       # → ./build/reports
make build_results       # → ./build/results
make build_housekeeping  # → ./build/housekeeping
make deploy              # build all + copy to ${HOME}/Testing/SkillSurvey

# Typical order of operations:
./survey            # crawl jobs
./reports           # compute monthly counts
./results           # start REST API (long-running)
./housekeeping cleanfs      # remove Chromium temp files
./housekeeping sendlog      # email error.log and clear it
```

## Error handling

- Errors written to `error.log` (next to executable) with timestamp and stack trace
- Panics recovered via `exception.ReportErrorIfPanic()` deferred in each binary's main
