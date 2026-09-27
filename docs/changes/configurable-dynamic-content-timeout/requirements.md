# Requirements: Configurable Dynamic Content Extraction Timeout and `runtask.json` Validation

## Context

`DynamicContentExtractor.ExtractDynamicContent` currently hard-codes a 60-second context timeout at
`runtask/internal/dynamiccontentextractor/dynamiccontentextractor.go:56`. That value governs every
chromedp page load in both the Seek and Jora adapters and cannot be changed without a rebuild.

This is the first step of the parent change recorded in
[`../use-vmm-guest-parent/preproposal.md`](../use-vmm-guest-parent/preproposal.md): running the
scraper inside a single-vCPU vmm guest is expected to make page loads slower, and a fixed 60 seconds
is identified there as the first thing likely to break. Making the timeout configurable is also
useful on its own, independently of whether the vmm move ever happens.

**Decision requiring agreement before `design.md`:** the timeout is specified below as a single
global value in `runtask.json`, applying to both adapters. The alternative — a per-site value in the
Seek and Jora adapter config files, alongside Jora's existing `SecondsBetweenJobPosts` and
`SecondsBetweenLinkPages` — is rejected here because the motivation is environmental (host speed),
not per-site, and a per-site knob would be a speculative abstraction.

## Timeout Configuration

### Reading the timeout from runtask.json
WHEN `runtask` starts and `runtask.json` contains `DynamicContentExtractionTimeoutSeconds` set to a positive integer THE SYSTEM SHALL use that value, in seconds, as the dynamic content extraction timeout.
WHEN the configured timeout is used THE SYSTEM SHALL apply the same value to both the Seek adapter and the Jora adapter.

### Defaulting when unconfigured
WHEN `runtask.json` does not contain a `DynamicContentExtractionTimeoutSeconds` field THE SYSTEM SHALL use a default timeout of 60 seconds.
WHEN `runtask.json` sets `DynamicContentExtractionTimeoutSeconds` to `0` THE SYSTEM SHALL treat the field as unset and use the default timeout of 60 seconds.

The absent case and the explicit-zero case are required to behave identically because decoding JSON
into a plain `int` field cannot distinguish them.

### Rejecting invalid values
WHEN `runtask.json` sets `DynamicContentExtractionTimeoutSeconds` to a negative number THE SYSTEM SHALL fail at startup with an error naming the field and the offending value.
WHEN the timeout value is rejected THE SYSTEM SHALL NOT run the requested command.

Failing fast is required here because a negative duration produces a context that is already expired,
which would surface as every page load failing for an unrelated-looking reason.

## Configuration Validation

Once one field is validated at load time it is inconsistent to leave the rest unchecked, so the same
treatment is extended to every field of `runtask.json` that has an invariant `runtask` can actually
check.

### What is and is not validated
WHEN `runtask` loads `runtask.json` THE SYSTEM SHALL validate every field for which a checkable invariant is stated below.
WHEN a field is an empty string or zero THE SYSTEM SHALL treat it as not configured and skip that field's format checks, unless a requirement below states otherwise.
WHEN a field has no invariant stated below THE SYSTEM SHALL accept any value for it.

Presence cannot be validated globally, because which fields are required depends on the command:
`housekeeping cleanfs` needs only `ErrorLogFile`, `housekeeping sendlog` additionally needs the SMTP
fields, and `scrape` and `report` need `PocketBaseUrl` and the service-account credentials.
`cmd/runtask/main.go` loads the config before it dispatches the command, so demanding every field be
populated would reject configurations that are valid for the command being run.

### Reporting validation failures
WHEN any validation check fails THE SYSTEM SHALL fail at startup with an error naming the field and the offending value.
WHEN a validation check fails THE SYSTEM SHALL NOT run the requested command.
WHEN more than one field fails validation THE SYSTEM SHALL report every failing field in a single error, not only the first.

Reporting all failures at once matters because `runtask.json` is hand-edited on a machine reached over
SSH; a one-error-per-run cycle is needlessly slow to work through.

### Error log file
WHEN `ErrorLogFile` is empty or absent THE SYSTEM SHALL fail at startup with an error naming the field.

`ErrorLogFile` is the one universally required field: `exception.Init(cfg.ErrorLogFile)` runs for
every command at `cmd/runtask/main.go:31` and already fails on an empty path inside `os.OpenFile`.
Validating it at load time changes no outcome, only the quality of the message.

### PocketBase URL
WHEN `PocketBaseUrl` is non-empty THE SYSTEM SHALL require that it parses as an absolute URL with an `http` or `https` scheme and a non-empty host.
WHEN `PocketBaseUrl` is non-empty and fails that check THE SYSTEM SHALL fail at startup with an error naming the field and the value.

### SMTP port
WHEN `SmtpPort` is non-zero THE SYSTEM SHALL require that it is between 1 and 65535 inclusive.
WHEN `SmtpPort` is zero or absent THE SYSTEM SHALL treat it as not configured and SHALL NOT fail validation.

### Email addresses
WHEN `ServiceAccountEmail` is non-empty THE SYSTEM SHALL require that it parses as a single RFC 5322 address.
WHEN `SenderEmail` is non-empty THE SYSTEM SHALL require that it parses as a single RFC 5322 address.
WHEN `EmailRecipient` is non-empty THE SYSTEM SHALL require that it parses as a single RFC 5322 address.

### Fields deliberately left unvalidated

| Field | Why no check |
|---|---|
| `ServiceAccountPassword` | No invariant `runtask` owns; PocketBase is the authority, and failure surfaces as an auth error |
| `SenderEmailPassword` | Same — the SMTP server is the authority |
| `SmtpDomain` | A hostname's only meaningful test is resolution, which is a network call and not a load-time concern |
| `SeekConfigFile` | Validated where it is opened, and a deployment may legitimately configure only one site |
| `JoraConfigFile` | Same as `SeekConfigFile` |

Checking that the two adapter config paths exist would make startup fail for a command that never
reads them, and would turn a per-site skip — which `scrape.adapterForSite` already logs and continues
past — into a total failure.

## Dynamic Content Extraction

### Applying the timeout
WHEN the extractor begins extracting content from a page THE SYSTEM SHALL abort that page's extraction once the configured timeout elapses.
WHEN the configured timeout elapses during a page extraction THE SYSTEM SHALL return a timeout error for that page.

The timeout is per `ExtractDynamicContent` call — that is, per page load — not a budget for an
entire `scrape` run or for a whole site.

### Error handling during a run
WHEN a page extraction fails with a timeout error THE SYSTEM SHALL log the error and continue with the remaining pages and job posts.
WHEN a page extraction fails with a timeout error THE SYSTEM SHALL still upsert the job posts that were successfully extracted before the failure.

## Unchanged Behaviours

WHEN `runtask` loads its configuration THE SYSTEM SHALL CONTINUE TO read `runtask.json` from the directory containing the executable, with no environment-variable fallback.
WHEN an existing deployed `runtask.json` that predates this change is loaded THE SYSTEM SHALL CONTINUE TO start successfully and scrape with the previous 60-second behaviour.
WHEN the `runtask.json` currently deployed in the dev checkout is loaded THE SYSTEM SHALL CONTINUE TO pass validation — it carries all eleven existing fields, a `PocketBaseUrl` with an `http://` scheme, and `SmtpPort` 587.
WHEN the extractor launches Chromium THE SYSTEM SHALL CONTINUE TO apply the existing user agent, 1920x1080 window size, headless flag, `disable-blink-features=AutomationControlled` flag, and `XDG_RUNTIME_DIR` fallback.
WHEN the extractor runs its chromedp actions THE SYSTEM SHALL CONTINUE TO evaluate the `navigator.webdriver` and `navigator.plugins` overrides before navigating.
WHEN `chromedp.Nodes` is called THE SYSTEM SHALL CONTINUE TO pass `chromedp.AtLeast(0)` so a missing selector does not block forever.
WHEN the Jora adapter paces its requests THE SYSTEM SHALL CONTINUE TO honour `SecondsBetweenJobPosts` and `SecondsBetweenLinkPages` from the Jora adapter config, unchanged.
WHEN `runtask housekeeping cleanfs` or `runtask housekeeping sendlog` runs THE SYSTEM SHALL CONTINUE TO work without PocketBase authentication and without consulting the timeout value.
WHEN `runtask scrape` completes THE SYSTEM SHALL CONTINUE TO upsert job posts idempotently, as covered by `TestScrapeRunIsIdempotent`.

## Out of Scope

Per-site timeout values in the Seek and Jora adapter config files.
Making any other hard-coded duration configurable, including Jora's inter-request pacing and the chromedp allocator options.
Any vmm guest creation or host configuration — that remains with the parent change.
Retrying a page extraction that timed out.
Per-command presence validation — checking that the fields a given command needs are populated, rather than only that populated fields are well-formed.
Load-time existence checks on `SeekConfigFile` and `JoraConfigFile`.
Any validation requiring network access, such as resolving `SmtpDomain` or contacting `PocketBaseUrl`.
Validation of the Seek and Jora adapter config files' own contents.
