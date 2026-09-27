# Pre-proposal: Run runtask and pocketbaseserver in separate OpenBSD vmm guests

Status: pre-proposal. This document precedes `requirements.md` — it records the survey of the
existing host, the proposed phasing, and the known complications, so that the requirements can be
written against facts rather than assumptions.

Server state surveyed 27 Sep 2026 against `host.t420` (OpenBSD 7.9) via read-only inspection.

## Goal

Isolate `runtask` (scrape / report / housekeeping) from `pocketbaseserver` (auth, collections,
REST API, frontend) by running each in its own OpenBSD vmm guest, rather than side by side in the
same host userland.

## What already exists on the host

This is largely a "fill in the gaps" job rather than greenfield work. The host already runs vmm
guests today.

| Item | State |
|---|---|
| Virtualisation | `vmm0 at mainbus0: VMX/EPT`; `vmd` and `vmctl` installed; `vmd` enabled via `rcctl` |
| Guest switch | `veb0` (description `switch8-vsubnet`) with members `vport0` and `tap0` |
| Guest gateway | `vport0` = `10.0.0.1/24`, configured from `/etc/hostname.vport0` |
| Guest DHCP | Host `dhcpd` on `vport0`; `/etc/dhcpd.conf` serves `10.0.0.2–254`, router `10.0.0.1`, DNS `192.168.8.1` |
| Routing | `net.inet.ip.forwarding=1` |
| VM definitions | `/etc/vm.conf` includes `/home/junying/VMs/vm.conf`, which includes one `vm.conf` per VM directory |
| Existing VMs | `pgserver` (running, 4G max / 2.1G current), `vm1` (stopped), `alpinedev` (stopped) |
| Guest template | `/home/junying/VMs/openbsdVmTemplate/` — 2.3G OpenBSD 7.9 disk image plus `install79.iso` |

Host resources:

| Resource | Value |
|---|---|
| CPU | Intel i5-2520M, `hw.ncpuonline=4` (2 physical cores) |
| RAM | 8 GB (`hw.physmem=8451125248`) |
| Swap | 16 GB on `/dev/sd1b` |
| Free disk | 48 GB on `/home`, 86 GB on `/usr/local` |
| LAN interface | `iwn0` (WiFi), carrying `192.168.8.143` plus aliases `.141`–`.154` |
| Wired interface | `em0` — **no carrier** |

## Proposed phasing

### Phase 0 — spec and unknowns

1. Per `CLAUDE.md`, agree `requirements.md` for this change before any host or guest configuration
   is touched. This is squarely "complex and costly to get wrong".
2. Establish where production `runtask` actually runs today. Every SkillSurvey line in `junying`'s
   crontab is commented out, and `/home/skillsurvey`'s crontab is not readable as `junying`. Until
   this is known, the change cannot state what it is decommissioning.

### Phase 1 — runtask guest (pocketbaseserver stays on the host)

3. Clone `openbsdVmTemplate/disk.qcow2` to `VMs/runtask/disk.qcow2`; write `VMs/runtask/vm.conf`
   declaring memory, disk, `owner junying:junying`, and `interface { switch "vsubnet" }`; add the
   include to `VMs/vm.conf`; start via `vmctl`.
4. Inside the guest: assign a static `10.0.0.x` (or a DHCP reservation), `pkg_add chromium`, create
   the service user, and size `/tmp` for Chromium's temporary directories.
5. Build `runtask` on the host (`cd runtask && make build`) and copy the binary plus a
   guest-specific `runtask.json` into the guest, rather than installing the Go toolchain into a
   single-vCPU guest.
6. Set that `runtask.json`'s `PocketBaseUrl` to `192.168.8.143:8090` — reachable from the guest via
   the `10.0.0.1` gateway — and confirm pf passes `10.0.0.0/24` to that address and NATs outbound
   traffic to Seek, Jora, and the SMTP relay.
7. Move the cron schedule (`scrape`, `report`, `housekeeping cleanfs`, `housekeeping sendlog`) into
   the guest and disable it on the host. `cleanfs` then cleans the guest's `/tmp`, which is correct —
   that is where Chromium's leftovers will accumulate.
8. Run runtask's integration tests wherever the requirements decide they live (see complications),
   then perform a real `scrape` against dev before cutting production over.

### Phase 2 — pocketbaseserver guest (optional; see recommendation)

9. Second guest on the same pattern, plus migration of `pb_data` and republication of `pb_public/`.
10. pf `rdr-to` so `192.168.8.143:8090` lands on the guest; adjust the frontend `.env` and the dev
    stack addresses; redo backups against the new location.

## Complications

Ordered worst first.

### Guests cannot hold a `192.168.8.x` address

`em0` has no carrier, so the LAN lives on `iwn0`, a WiFi interface. A vmm tap cannot be bridged onto
an 802.11 station interface, which is why `veb0` is a NAT'd `10.0.0.0/24` island rather than a
bridge onto the LAN. Any guest is therefore reachable from the LAN only through a pf `rdr-to` on the
host.

This is harmless for runtask, which is outbound-only. For `pocketbaseserver` it means production's
`192.168.8.143:8090` becomes a redirect through the host, so the host remains a single point of
failure and the isolation gained is thinner than it appears.

### pf changes require root at the console

`/etc/pf.conf` is `0600 root:wheel` and `doas` is not passwordless on this host. Every rule change —
passing traffic for the new guest, NAT, and later `rdr-to` — is a manual, deliberate act. This
matches the existing posture around production restarts and should be budgeted for, not routed
around.

### Memory is the binding constraint

8 GB total, with `pgserver` already reserving 4 GB. Headless Chromium wants roughly 1–2 GB by
itself. A 2 GB runtask guest plus a 1 GB PocketBase guest leaves very little host headroom, so
`pgserver` would likely need shrinking. The 16 GB of swap does not help here: swapping during a
scrape is worse than not isolating it.

### One vCPU per guest, and chromedp is already the fragile part

Every VM reports `VCPUS 1`. Scrape tests already produce chromedp websocket timeouts when the host
is busy. Moving Chromium into a single-vCPU guest backed by qcow2 I/O makes that strictly worse. The
hard-coded 60-second timeout in `runtask/internal/dynamiccontentextractor/dynamiccontentextractor.go:57`
is the first thing likely to bite.

### The guest template's `vm.conf` is booby-trapped

`/home/junying/VMs/openbsdVmTemplate/vm.conf` still declares `vm "pgserver"` pointing at
`/home/junying/VMs/pgserver/disk.qcow2`. Copying it verbatim produces a duplicate VM name and a
second VM attached to the running pgserver's disk. The name and disk path must be rewritten before
the file is included.

### Guest `/tmp` sizing

A default OpenBSD install gives `/tmp` a few hundred megabytes. Chromium temp directories are
precisely what `housekeeping cleanfs` exists to remove, which indicates they accumulate in practice.
The guest needs a deliberately large `/tmp` (the host's is 3.9 GB) or `cleanfs` becomes load-bearing
rather than tidy-up.

### Configuration and secrets fan out

`config.Load()` in `runtask/internal/config/config.go` reads `runtask.json` from the binary's own
directory, with no environment-variable fallback. The guest therefore needs its own copy, carrying
the service-account password and the SMTP password. Those credentials would then exist on a second
filesystem, inside a qcow2 image under `/home/junying`. How that image is protected and backed up
needs a deliberate decision.

### SMTP egress

`housekeeping sendlog` dials out through `net/smtp`. That path now crosses NAT. If the current setup
depends on anything host-local, it will fail quietly into the error log — which also now lives
inside the guest.

### Where tests run

`CLAUDE-project.md` mandates that all tests run on the OpenBSD server, with integration tests
starting a real PocketBase on a `t.TempDir()` data directory. Once runtask's runtime moves into a
guest there are two options:

- keep running tests in the `junying` host checkout — fast, but no longer the real environment, and
  the flakiness that matters here is environmental; or
- install Go and chromium in the guest and make it the test host — accurate, but slow on one vCPU.

This should be decided in `requirements.md`, not discovered during implementation.

### Documentation and tooling drift

`CLAUDE-project.md`'s deployment steps, `CLAUDE.local.md`'s addresses, and the `/openbsd-run` skill
all assume both binaries live in the `junying` host checkout. All three need updating, or later work
will deploy to the wrong place.

### Boot ordering

`vmd` brings guests up independently, so runtask's cron can fire before PocketBase is serving. Today
the two share a machine and this mostly works by luck. `scrape` and `report` need a retry or a
readiness check.

### Clock drift

vmm guests drift. PocketBase auth tokens and the month bucketing in `report` both depend on the
clock. `ntpd` must be confirmed enabled in each guest.

## Recommendation

Stop after Phase 1.

`runtask` is the stateless, resource-hungry, crash-prone half. It is nearly free to isolate, and
that is where the benefit of this change lives.

`pocketbaseserver` is stateful, LAN-facing, and pinned to `192.168.8.143:8090`. Moving it behind a
pf redirect on the same host buys isolation that is unlikely to be felt, at the cost of a live
production data migration.

If Phase 2 is still wanted afterwards, it should be raised as its own change with its own
`requirements.md`.

## Open questions for `requirements.md`

1. Where does production `runtask` run today, and under which user and schedule?
2. Do runtask's tests run in the guest or stay in the host checkout?
3. What memory budget does the runtask guest get, and does `pgserver` shrink to make room?
4. Is the guest's qcow2 image — containing the service-account and SMTP credentials — in scope for
   backup, and if so, to where?
5. Is Phase 2 in scope for this change at all, or deferred to a separate one?
