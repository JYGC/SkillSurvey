# Deploying SkillSurvey

SkillSurvey runs on two kinds of OpenBSD host:

| Host | Runs | Installed by |
|---|---|---|
| App host | `pocketbaseserver` on port 8090, and the frontend served by `httpd` on port 80 | `install-server.sh` from a server bundle |
| Runtask host | `runtask`, Chromium and the job schedule | `install-runtask.sh` from a runtask bundle |

Every deployment is manual. CI only builds the bundles. You copy them to each host and run the install script as
root. Both scripts can be run as often as you like: a second run with the same bundle and options changes nothing and
ends with `0 changes`.

The commands below use these placeholders:

| Placeholder | Means |
|---|---|
| `<ci-host>` | The OpenBSD host that runs CI |
| `<user>` | Your login on that host, which can read CI's run logs |
| `<login>` | Your login on the app or runtask host, which can use `doas` |
| `<app-host>` | The app host's name, as browsers will reach it |
| `<runtask-host>` | The runtask host's name |
| `<short>` | The first 7 characters of the commit being deployed |

## Getting the bundles

Each passing CI run on `master` ends with `build-bundles`, which leaves two files in that run's artifact folder:

```text
/var/ci/logs/SkillSurvey/<UTC time>-<short>/artifacts/skillsurvey-server-<short>.tgz
/var/ci/logs/SkillSurvey/<UTC time>-<short>/artifacts/skillsurvey-runtask-<short>.tgz
```

CI keeps them for 30 days. Copy them to your workstation once, then install that same copy everywhere, so test and
production run exactly the same files:

```sh
# On your workstation, once per commit
mkdir -p ~/skillsurvey-bundles
scp '<user>@<ci-host>:/var/ci/logs/SkillSurvey/*-<short>/artifacts/*.tgz' ~/skillsurvey-bundles/

# For each host
scp ~/skillsurvey-bundles/skillsurvey-server-<short>.tgz <login>@<app-host>:
scp ~/skillsurvey-bundles/skillsurvey-runtask-<short>.tgz <login>@<runtask-host>:
```

### Building a bundle by hand

If a commit's bundles have expired, build them again on the CI host from your own checkout:

```sh
git checkout <commit>
sh deploy/build-bundles.sh
```

The bundles appear in `skillsurvey-bundles/` beside the checkout. The script builds an export of the commit, so
uncommitted changes are left out, and it says so. It keeps bundles that already exist rather than building them
again. The programs are byte-identical to CI's if Go has not changed in between.

### What a bundle holds

```text
skillsurvey-server-<short>/          skillsurvey-runtask-<short>/
  BUNDLE-INFO                          BUNDLE-INFO
  SHA256                               SHA256
  install-server.sh                    install-runtask.sh
  install-common.sh                    install-common.sh
  pocketbaseserver.rc                  run-scheduled-job
  pocketbaseserver                     runtask
  frontend/                            examples/
```

`BUNDLE-INFO` names the commit and the OpenBSD release and architecture it was built on. The scripts refuse to run on
any other release or architecture, and refuse if any file fails its check against `SHA256`.

## Installing an app host

```sh
tar -xzf skillsurvey-server-<short>.tgz
cd skillsurvey-server-<short>
doas sh install-server.sh --listen-address 0.0.0.0 \
  --site-address http://<app-host> \
  --pocketbase-address http://<app-host>:8090
```

| Option | Means |
|---|---|
| `--listen-address` | The address both services listen on. `0.0.0.0` means every address, which suits a host whose address comes from DHCP |
| `--site-address` | Where browsers open the site. PocketBase allows cross-origin calls from this address only |
| `--pocketbase-address` | Where browsers reach PocketBase. It is written to `host-config.json`, which the frontend reads at startup |

All three are needed the first time. They are saved in `/etc/skillsurvey/server.conf`, so later runs can leave them
out. Giving a new value changes just that value: a new listen or site address restarts PocketBase, and a new
PocketBase address rewrites `host-config.json` only.

The script creates the `_skillsurvey` account, installs the program and the service definition, writes
`/etc/httpd.conf`, enables and starts `pocketbaseserver` and `httpd`, and checks that both answer. It installs no
packages. It refuses to touch an `/etc/httpd.conf` it did not write.

### Steps left by hand after a first install

The script prints these too:

1. Create the first superuser. The password shows in `ps` for a moment:

   ```sh
   doas -u _skillsurvey /usr/local/libexec/skillsurvey/pocketbaseserver superuser upsert EMAIL PASSWORD --dir=/var/skillsurvey/pb_data
   ```

2. In the dashboard at `http://<app-host>:8090/_/`, set a password on `runtask@skillsurvey.com`. The runtask host
   needs it.
3. Open `http://<app-host>` and `http://<app-host>:8090/_/` from another machine. Only that shows whether the
   firewall lets them through.

## Installing a runtask host

```sh
tar -xzf skillsurvey-runtask-<short>.tgz
cd skillsurvey-runtask-<short>
doas sh install-runtask.sh
```

The script creates the `_runtask` account and its `runtask` login class, installs the `chromium` package, installs
the program, and creates any missing settings file in `/etc/runtask/` from its example. It never overwrites a
settings file that exists. It refuses if the X libraries Chromium needs are missing, so install the X sets first.

### Steps left by hand after a first install

1. Replace the two site settings files with the real ones:

   ```sh
   doas install -o root -g _runtask -m 0640 www.seek.com.au.json au.jora.com.json /etc/runtask/
   ```

2. Fill in `/etc/runtask/runtask.json` with `doas vi`, using the password set on `runtask@skillsurvey.com`. Every
   `REPLACE_ME` must go, including the mail settings, because `housekeeping sendlog` runs every week.
3. Start the schedule, giving the hours at which to scrape:

   ```sh
   doas sh install-runtask.sh --start-schedule --scrape-hours 7,17
   ```

Before writing the schedule, the script refuses if any settings file still holds `REPLACE_ME`, and runs
`runtask check-setup` as `_runtask` under its login class. That check loads both site files, signs in to PocketBase,
reads `sites` and starts Chromium, and it writes nothing. You can run it yourself at any time:

```sh
doas su -c runtask -s /bin/sh _runtask -c '/usr/local/libexec/runtask/run-scheduled-job check-setup'
```

### The schedule

The schedule lives in `/etc/crontab` between two marker lines, and `cron` picks up a change by itself:

```text
# BEGIN runtask schedule. install-runtask.sh replaces everything down to the END line.
MAILTO=root
0 7,17 * * * _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job scrape
0 1 1 * * _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job report
0 19 * * 5 _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job housekeeping cleanfs
0 19 * * 5 _runtask -n -s /usr/local/libexec/runtask/run-scheduled-job housekeeping sendlog
# END runtask schedule
```

Times are the host's local time, which the script prints. `-n` sends no mail after a job succeeds, and `-s` stops a
job starting while its last run is still going. Running the script without `--start-schedule` never touches the
schedule. Running it with other hours replaces the block, leaving the rest of `/etc/crontab` alone.

Jobs write errors to `/var/log/runtask/error.log`, which `sendlog` emails and empties every Friday. A job that
crashes leaves mail for root on the host.

## Upgrading

Copy the new bundles across as above, unpack them, and run the same script with no options:

```sh
doas sh install-server.sh
doas sh install-runtask.sh
```

On an app host:

- If `pocketbaseserver` differs, the script stops it, copies `/var/skillsurvey/pb_data` to
  `/var/skillsurvey/snapshots/<UTC time>-<old short>/`, replaces the program, starts it so its migrations run, and
  waits up to 60 seconds for it to answer. It keeps the 3 newest snapshots it made. It refuses before stopping
  anything if `/var` lacks room for the snapshot.
- If only the frontend differs, PocketBase keeps running. The new files are unpacked beside the live ones, and
  `current` is switched to them in one step. The previous release is kept.

On a runtask host, the program is replaced by renaming, so a job already running finishes on the old one, and the
next job uses the new one.

## Going back after a failed upgrade

If PocketBase does not answer after an upgrade, the script stops, puts nothing back, and prints these commands with
the real snapshot name. `pb_data` is set aside rather than deleted, so the failed state can still be examined:

```sh
doas rcctl stop pocketbaseserver
doas mv /var/skillsurvey/pb_data /var/skillsurvey/pb_data.failed
doas cp -Rp /var/skillsurvey/snapshots/<name>/pb_data /var/skillsurvey/pb_data
doas sh <the previous bundle>/install-server.sh
```

To go back on the frontend only, run the previous bundle's `install-server.sh`. To go back on a runtask host, run the
previous bundle's `install-runtask.sh`.

## Where things are

### App host

| Path | Holds |
|---|---|
| `/usr/local/libexec/skillsurvey/pocketbaseserver` | The program |
| `/etc/rc.d/pocketbaseserver` | The service definition. Its flags are kept by `rcctl` |
| `/etc/skillsurvey/server.conf` | The three addresses |
| `/etc/httpd.conf` | Written whole by the script, with a marker on its first line |
| `/var/skillsurvey/pb_data/` | The data. No script ever removes or moves it |
| `/var/skillsurvey/snapshots/` | Copies of `pb_data` from before each program upgrade. They hold password hashes, so only root can read them |
| `/var/www/htdocs/skillsurvey/current` | A link to the live release, with `previous` beside it |
| `/var/db/skillsurvey/BUNDLE-INFO` | The installed bundle |
| `/var/log/daemon` | PocketBase's log, through syslog |

### Runtask host

| Path | Holds |
|---|---|
| `/usr/local/libexec/runtask/runtask` | The program |
| `/usr/local/libexec/runtask/run-scheduled-job` | Sets `HOME`, `PATH` and the working folder, for cron and `check-setup` alike |
| `/usr/local/libexec/runtask/runtask.json` | A link to `/etc/runtask/runtask.json`, because `runtask` looks for its settings beside itself |
| `/etc/runtask/` | `runtask.json` and the two site files, readable by root and `_runtask` only |
| `/etc/login.conf.d/runtask` | The login class, with room for Chromium |
| `/var/runtask/` | `_runtask`'s home, where Chromium keeps its folders |
| `/var/log/runtask/error.log` | Job errors |
| `/var/db/runtask/BUNDLE-INFO` | The installed bundle |

## Testing the scripts

`go test ./deploy/...` runs each script against a throwaway root, with stubs standing in for the commands that need
root. It needs OpenBSD's `sha256`, so it runs in CI and skips elsewhere. Nothing under `deploy/` may name a real host
or a private address, because the repository is public, and a test enforces that.
