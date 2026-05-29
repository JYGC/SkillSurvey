![Alt Text](https://media.giphy.com/media/oz8HuJkUQZEv2ylE7u/giphy.gif)

# SkillSurvey

SkillSurvey watches Australian job boards — Seek and Jora — and tracks how often each technical skill appears in job listings month by month. The goal is to give a clear picture of which skills employers are actually asking for, and how that demand is changing over time.

## What it does

Every month, SkillSurvey:

1. **Collects job listings** from Seek and Jora automatically.
2. **Counts skill mentions** — for each skill it tracks (e.g. Python, Kubernetes, React), it counts how many listings mentioned that skill during the month.
3. **Publishes the results** through a web dashboard showing a chart of monthly counts per skill over the past year.

## Who it is for

Anyone curious about the Australian tech job market — developers planning what to learn next, hiring managers benchmarking against the market, or researchers studying industry trends.

## Server requirements (OpenBSD)

Install the following system packages once on the server:

```sh
pkg_add go node chromium chromedriver
```

| Package | Why it is needed |
|---------|-----------------|
| `go` | Compiles the database server and the scheduled task runner |
| `node` | Runs the frontend build tools |
| `chromium` | Headless browser used by the scraper to load job listing pages |
| `chromedriver` | Controls Chromium during automated end-to-end tests |

No global npm packages are required — all frontend tooling is installed locally via `npm install` inside the `frontend/` directory.

## Parts of the system

| Part | What it does |
|------|-------------|
| **Web scraper** | Visits Seek and Jora on a schedule and saves job listings |
| **Report generator** | Counts skill mentions in the saved listings and stores monthly totals |
| **Web interface** | Shows the monthly totals as an interactive chart |
| **Database** | Stores listings, skill definitions, and monthly counts |
