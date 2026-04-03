# Production environment
- Operation System: OpenBSD
  - Ensure all projects can run on OpenBSD
  - OpenBSD system has chrome already installed for chromedp to work

# New Tech stack
## pocketbaseserver
- Language: Go
- Framework: PocketBase
## runtask
- Language: Go
- Web Scraping: Chromedp (headless Chrome automation)
- Job Boards: Seek and Jora adapters
## migrate
- Language: Go
- Old DB Client: GORM (reading from legacy SQLite)
- New DB Client: github.com/r--w/pocketbase (writing to pocketbaseserver)
- Temporary: deleted after one-shot data migration
- Unit tests must be written before implementation
## frontend
- Framework: Vue 3 with Vue Router
- Language: TypeScript
- UI Libraries: IBM Carbon Design System, Bootstrap Vue 3
- Charts: Chart.js
- Backend Client: PocketBase JS SDK
- Build Tool: Vite

# Go code style
- Use styling described here: https://google.github.io/styleguide/go/guide
- Format all Go code with `gofmt` and `goimports` before committing

# Typescript code style
- Use styling described here: https://google.github.io/styleguide/tsguide.html
- Run `npm run lint` in `frontend/` to check and fix frontend code style

# Vue code style
- Use styling described here: https://vuejs.org/style-guide/
- Use composition API

# Workflow rules
- Always read a file before editing it
- Create a plan before implementing complex features
- Build `pocketbaseserver` with `make build`; run in dev mode with `make run_dev` (from `pocketbaseserver/`)
- Build `runtask` with `make build`; run in dev mode with `make run_dev` (from `runtask/`)

# Database rules
- All schema changes must be made via migrations in `pocketbaseserver/migrations/`, never by editing the database directly

# Testing rules
- All new features require integration tests
- Do not mock the database in integration tests
- Go tests use the standard `testing` package against a real PocketBase test instance
- Frontend unit tests are not required, but integration tests for new API-connected features are

# Quality
- No speculative abstractions — only build what's needed now
- Integration Tests must be created first before actual code

# Old Tech stack
## backend
- Language: Go
- Web Scraping: Chromedp (headless Chrome automation)
- ORM: GORM
- Database: SQLite
- HTTP: Standard Go HTTP server with CORS support
- Job Boards: Seek and Jora adapters
## frontend
- Framework: Vue 3 with Vue Router
- Language: TypeScript
- UI Libraries: IBM Carbon Design System, Bootstrap Vue 3
- Charts: Chart.js
- Backend Client: PocketBase JS SDK
- Build Tool: Vite

# Security
- Never commit secrets or API keys