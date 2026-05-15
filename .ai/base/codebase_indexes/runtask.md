# runtask — Codebase Index

## Purpose

Command-line task runner that handles periodic scraping, reporting, and maintenance. Authenticates with PocketBase as a service account and executes scheduled jobs.

## Tech stack

- Go 1.24.1
- PocketBase client (`github.com/r--w/pocketbase`)
- Chromedp (headless Chrome) for dynamic-page scraping (Seek detail pages, Jora)
- Target OS: OpenBSD

## Directory map

```
runtask/
├── cmd/runtask/
│   └── main.go                         # Command dispatcher (scrape | report | housekeeping)
├── internal/
│   ├── config/
│   │   └── config.go                   # JSON config loader (runtask.json)
│   ├── dynamiccontentextractor/
│   │   └── dynamiccontentextractor.go  # Chromedp wrapper: user-agent, wait, AtLeast(0)
│   ├── exception/
│   │   ├── errorlogging.go             # Global ErrorLogger; nil-safe helpers
│   │   └── errormessages.go            # Shared error message strings
│   ├── housekeeping/
│   │   ├── housekeeping.go             # CleanFS + SendLog
│   │   └── housekeeping_test.go
│   ├── pbclient/
│   │   └── pbclient.go                 # Typed PocketBase client wrapper
│   ├── report/
│   │   ├── report.go                   # Monthly count report generation
│   │   └── report_test.go
│   ├── scrape/
│   │   ├── scrape.go                   # Orchestrates site adapters → PocketBase upserts
│   │   └── scrape_test.go
│   └── siteadapters/
│       ├── siteadapterbase.go          # ISiteAdapter interface + JobPostResult type
│       ├── seekadapter.go              # Seek search API + Chromedp detail pages
│       ├── seekadapterconfig.go        # Config struct for Seek adapter
│       ├── seekapigetparameters.go     # Builds Seek API query parameters
│       ├── joraadapter.go              # Jora Chromedp scraper
│       └── joraadapterconfig.go        # Config struct for Jora adapter
├── Makefile                            # build | run targets
├── go.mod                              # module: keybook/runtask
└── go.sum
```

## Config (runtask.json)

```json
{
  "PocketBaseUrl":          "http://192.168.8.147:8090",
  "ServiceAccountEmail":    "runtask@skillsurvey.com",
  "ServiceAccountPassword": "<password>",
  "SeekConfigFile":         "./au.seek.com.au.json",
  "JoraConfigFile":         "./au.jora.com.json",
  "ErrorLogFile":            "./error.log",
  "SmtpDomain":             "smtp.gmail.com",
  "SmtpPort":               587,
  "SenderEmail":            "<gmail>",
  "SenderEmailPassword":    "<app-password>",
  "EmailRecipient":         "<recipient>"
}
```

## Commands

| Command | Description |
|---|---|
| `runtask scrape` | Scrapes Seek and Jora; upserts jobPosts into PocketBase |
| `runtask report` | Reads jobPosts + skill aliases; writes monthlyCountReports |
| `runtask housekeeping cleanfs` | Removes Chromium temp dirs under `/tmp` |
| `runtask housekeeping sendlog` | Emails `ErrorLogFile` contents; truncates the file |

## Key packages

### config
Loads `runtask.json` from the directory containing the executable. No env vars.

### pbclient
Typed wrapper around `github.com/r--w/pocketbase`. Handles auth and provides:
- `GetSites()` — list of site records
- `UpsertJobPost()` — create or update by jobSiteNumber
- `GetAllSkillNamesWithAliases()` — skill names + aliases for report matching
- `UpsertMonthlyCountReport()` — create or update by identifier

### scrape
`scrape.Run(cfg, pb)` iterates over PocketBase sites, selects the matching adapter by site name, runs the survey, and upserts results.

Site adapter selection (in `adapterForSite`):
- Site name matches adapter config file base name (case-insensitive), or equals "seek"/"jora"

### report
`report.Run(cfg, pb)` reads all job posts and skill aliases, counts alias occurrences in job post bodies (word-boundary matching), and upserts monthly count reports.

### housekeeping
- `CleanFS(baseDir)` — removes `.org.chromium.Chromium.*` and `chromedp-runner*` dirs
- `SendLog(cfg)` — reads error log, sends via `smtp.PlainAuth` on `SmtpDomain:SmtpPort`, truncates log

### exception
Package-level `ErrorLogger *log.Logger`. Call `Init(path)` once at startup. All logging functions (`LogErrorWithLabel`, `LogExtraData`, `ReportErrorIfPanic`) are nil-safe and fall back to the standard logger if `Init` was not called (e.g. in tests).

### dynamiccontentextractor
Chromedp helper used by both adapters. Key detail: all `chromedp.Nodes` calls use `chromedp.AtLeast(0)` to prevent indefinite blocking when a selector is not found.

## Error handling

- Errors written to `ErrorLogFile` with timestamp and short file/line via `exception.ErrorLogger`
- Panics recovered via `exception.ReportErrorIfPanic()` deferred in main
- `housekeeping sendlog` emails the file to `EmailRecipient` then clears it

## Tests

Integration tests start a real PocketBase instance (`t.TempDir()` data dir) with all `pocketbaseserver/migrations` applied, then exercise the package against it. No mocking.

- `housekeeping_test.go` — real SMTP stub on a random TCP port
- `report_test.go` — seeds skill data and job posts via app internal API
- `scrape_test.go` — `httptest.Server` stub for Seek API + job detail pages
