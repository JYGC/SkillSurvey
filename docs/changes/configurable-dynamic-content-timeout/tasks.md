# Tasks: Configurable Dynamic Content Extraction Timeout and `runtask.json` Validation

Implements [`design.md`](./design.md) against [`requirements.md`](./requirements.md).

Tasks 1–3 are independent and come first: they are the test tasks, and per `CLAUDE.md` each is
written and seen to fail before the code it covers exists. Because Go will not compile a call to a
function signature that does not exist yet, the initial red state for tasks 1–3 is a **compile
failure**. That is the expected outcome — do not soften a test to fit the current signatures.

| # | Task | Required | Depends on |
|---|---|---|---|
| 1 | Integration test: configured timeout governs extraction | Required | — |
| 2 | Unit test: timeout resolution | Required | — |
| 3 | Unit test: config validation | Required | — |
| 4 | Implement config field, accessor, load split, validation | Required | 2, 3 |
| 5 | Thread the duration from scrape to the chromedp context | Required | 1, 4 |
| 6 | Regression run on the OpenBSD server | Required | 5 |
| 7 | Document the new field | Required | 4 |
| 8 | Verify the production `runtask.json` against the new rules | Required | 4 |
| 9 | Format with `gofmt` and `goimports` | Required | 5 |
| 10 | Set a longer timeout in the deployed config | Optional | 5 |

---

## 1. Integration test: configured timeout governs extraction

**Required.** Depends on nothing.

New file `runtask/internal/dynamiccontentextractor/dynamiccontentextractor_test.go`, following the
`httptest` stub pattern already used by `runtask/internal/scrape/scrape_test.go`.

- `TestExtractDynamicContentAbortsWhenConfiguredTimeoutElapses` — an `httptest` server whose handler
  sleeps roughly 30 seconds before responding; an extractor built with a 2-second timeout; assert
  `ExtractDynamicContent` returns a non-nil error **and** that the call returned in well under the
  handler's sleep (under 20 seconds). The elapsed-time assertion is the point: it is what proves the
  configured value governed the abort rather than some unrelated failure.
- `TestExtractDynamicContentSucceedsWithinGenerousConfiguredTimeout` — the same stub responding
  immediately with a small HTML document; an extractor built with a 30-second timeout; assert
  `ExtractDynamicContent` returns nil and that the extract function actually observed the page
  content.

**Expected outcome:** the package fails to build, because `NewDynamicContentExtractor` currently takes
no arguments. This is the red state.

**Note:** these tests drive real headless Chromium. That is not a new dependency —
`TestScrapeRunCreatesJobPosts` already does so through the Seek adapter's job-page fetch — but it does
mean they only run on the OpenBSD server, which has `chromium-147` installed.

## 2. Unit test: timeout resolution

**Required.** Depends on nothing.

New file `runtask/internal/config/config_test.go`, in package `config` rather than `config_test`, so
it can reach the unexported `loadFromFile`.

Table test over `loadFromFile`, each case writing a JSON file into `t.TempDir()`. Every case includes a
non-empty `ErrorLogFile`, since that field becomes required:

| Input | Expected |
|---|---|
| timeout field absent | `DynamicContentExtractionTimeout()` == 60s |
| `"DynamicContentExtractionTimeoutSeconds": 0` | 60s |
| `"DynamicContentExtractionTimeoutSeconds": 180` | 180s |
| `"DynamicContentExtractionTimeoutSeconds": -1` | `loadFromFile` returns an error naming the field |

Plus `TestZeroValueConfigResolvesToDefaultTimeout`, which touches no file at all: `Config{}` must
resolve to 60 seconds. This is the guard on the five struct literals listed in `design.md`.

**Expected outcome:** the package fails to build — the field, the accessor, and `loadFromFile` do not
exist yet.

## 3. Unit test: config validation

**Required.** Depends on nothing. Same file as task 2.

Table test over `Config.validate()` using struct literals, one case per rule in `design.md`:
well-formed config returns nil; empty `ErrorLogFile`; `PocketBaseUrl` without a scheme, without a host,
and with a non-HTTP scheme; empty `PocketBaseUrl` producing no error; `SmtpPort` 587, 70000, and 0;
a malformed `SenderEmail`; each email field empty producing no error for that field.

Two cases carry specific requirements and must not be dropped:

- **Two simultaneous failures** — empty `ErrorLogFile` together with `SmtpPort` 70000 must produce a
  single error naming **both** fields. This is the requirement that all failures are reported together.
- **A `cleanfs`-shaped config** — a `Config` holding only `ErrorLogFile` must validate successfully.
  This is the guard against validation quietly becoming a presence check on every field.

**Expected outcome:** the package fails to build — `validate` does not exist yet.

## 4. Implement config field, accessor, load split, validation

**Required.** Depends on tasks 2 and 3.

In `runtask/internal/config/config.go`:

- add `DynamicContentExtractionTimeoutSeconds int` to `Config`;
- add `const defaultDynamicContentExtractionTimeoutSeconds = 60`;
- add `func (c Config) DynamicContentExtractionTimeout() time.Duration`, resolving any non-positive
  value to the default;
- split `Load()` into the existing path resolution plus
  `loadFromFile(configFilePath string) (Config, error)` holding the open, decode, and validate steps;
- add `func (c Config) validate() error` accumulating into a slice and returning `errors.Join(...)`,
  plus the `validateHttpUrl` helper;
- call `validate()` from `loadFromFile` after decoding.

**Expected outcome:** tasks 2 and 3 pass. Task 1 still fails to compile.

## 5. Thread the duration from scrape to the chromedp context

**Required.** Depends on tasks 1 and 4.

These four files must change together — the module does not compile in between:

- `runtask/internal/dynamiccontentextractor/dynamiccontentextractor.go` — add the
  `dynamicContentExtractionTimeout time.Duration` field, take it as a constructor parameter, convert
  the positional composite literal at line 38 to a keyed literal, and use the field in place of
  `60*time.Second` at line 56.
- `runtask/internal/siteadapters/seekadapter.go:25` — `NewSeekAdapter` accepts the duration and
  forwards it.
- `runtask/internal/siteadapters/joraadapter.go:25` — `NewJoraAdapter` accepts the duration and
  forwards it.
- `runtask/internal/scrape/scrape.go:63` and `:65` — pass `cfg.DynamicContentExtractionTimeout()`.

Neither `siteadapters` nor `dynamiccontentextractor` may gain an import of `internal/config`.

**Expected outcome:** task 1 passes; `go build ./runtask/...` is clean.

## 6. Regression run on the OpenBSD server

**Required.** Depends on task 5.

Push, pull on the server, then run `go test ./runtask/... -count=1` on an otherwise idle machine.

`TestScrapeRunCreatesJobPosts` and `TestScrapeRunIsIdempotent` must pass **unmodified**. Needing to
edit either one is the signal that the defaulting landed in `Load()` instead of the accessor.

**Expected outcome:** the whole `runtask` module is green. chromedp websocket timeouts on this host are
load-related; re-run once idle before treating one as a real failure.

## 7. Document the new field

**Required.** Depends on task 4.

- `.ai/base/codebase_indexes/runtask.md` — add `"DynamicContentExtractionTimeoutSeconds": 60` to the
  sample config at lines 53–67, and note under `### config` (line 81) that the loader validates fields
  and that an absent timeout means 60 seconds.
- `CLAUDE-project.md` — extend the runtask configuration sentence at line 55 with the new field, its
  60-second default, and the fact that malformed values are now rejected at startup.

**Expected outcome:** both files describe the field and the validation behaviour.

## 8. Verify the production `runtask.json` against the new rules

**Required.** Depends on task 4. Manual, and a prerequisite for deployment rather than for the build.

Validation converts a previously tolerated malformed value into a startup failure, so the production
config must be checked before the new binary runs. As `skillsurvey`, confirm that `ErrorLogFile` is
non-empty, `PocketBaseUrl` has an `http`/`https` scheme and a host, `SmtpPort` is in range, the three
email fields parse as addresses, and any timeout value present is not negative.

This cannot be done as `junying` — `/home/skillsurvey/` is not readable. The dev copy at
`runtask/build/runtask.json` has already been checked and passes.

**Expected outcome:** a production config known to satisfy the new rules, or a corrected one.

## 9. Format with `gofmt` and `goimports`

**Required.** Depends on task 5. Per the development rules in `CLAUDE-project.md`, before committing.

**Expected outcome:** no diff from either tool.

## 10. Set a longer timeout in the deployed config

**Optional.** Depends on task 5.

Only meaningful once `runtask` actually runs inside the vmm guest, which is the parent change's work.
Until then the 60-second default preserves current behaviour and nothing needs setting.

Note that `runtask/build/runtask.json` is not tracked in git, so this is a server-side edit rather
than a repository change.

**Expected outcome:** deferred to [`../use-vmm-guest-parent/preproposal.md`](../use-vmm-guest-parent/preproposal.md).
