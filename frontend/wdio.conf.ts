import type { Options } from '@wdio/types';
import { spawn, execFileSync, ChildProcess } from 'child_process';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

const REPO_ROOT = path.resolve(__dirname, '..');
const PB_BIN = path.join(REPO_ROOT, 'pocketbaseserver', 'build', 'pocketbaseserver');
const PB_CWD = path.join(REPO_ROOT, 'pocketbaseserver');
const PB_HOST = '192.168.8.147';
const PB_PORT = '8090';
export const PB_URL = `http://${PB_HOST}:${PB_PORT}`;

const chromeBin = process.env.CHROMIUM_PATH ?? '/usr/local/bin/chrome';
const chromedriverBin = process.env.CHROMEDRIVER_PATH ?? '/usr/local/bin/chromedriver';
const CHROMEDRIVER_PORT = 9515;

let chromedriverProcess: ChildProcess | null = null;
let pbServer: ChildProcess | null = null;
let tmpDir: string | null = null;

async function waitForUrl(url: string, timeoutMs = 15_000): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(url);
      if (res.ok) return;
    } catch {
      // not ready yet
    }
    await new Promise((r) => setTimeout(r, 200));
  }
  throw new Error(`Timed out waiting for ${url}`);
}

export const config: Options.Testrunner = {
  runner: 'local',
  specs: ['./tests/e2e/**/*.e2e.ts'],
  framework: 'mocha',
  reporters: ['spec'],
  mochaOpts: { timeout: 30000 },
  hostname: 'localhost',
  port: CHROMEDRIVER_PORT,
  path: '/',
  capabilities: [
    {
      browserName: 'chrome',
      'goog:chromeOptions': {
        binary: chromeBin,
        args: ['--headless', '--no-sandbox', '--disable-dev-shm-usage'],
      },
      'goog:loggingPrefs': { browser: 'ALL' },
    },
  ],
  baseUrl: PB_URL,

  async onPrepare() {
    // Kill any leftover pocketbaseserver process on this port
    try { execFileSync('pkill', ['pocketbaseserver'], { stdio: 'pipe' }); } catch { /* none running */ }
    await new Promise((r) => setTimeout(r, 500));

    // Start a fresh test PocketBase instance
    tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'pb-e2e-'));

    const { SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD } = await import('./tests/setup/seed');

    execFileSync(
      PB_BIN,
      ['superuser', 'upsert', SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD, '--dir', tmpDir],
      { stdio: 'pipe' },
    );

    pbServer = spawn(PB_BIN, ['serve', '--http', `${PB_HOST}:${PB_PORT}`, '--dir', tmpDir], {
      cwd: PB_CWD,
      stdio: 'pipe',
    });

    pbServer.on('error', (err) => { throw new Error(`pocketbaseserver failed: ${err.message}`); });

    await waitForUrl(`${PB_URL}/api/health`);

    process.env.TEST_PB_URL = PB_URL;
    const { seedInitialData } = await import('./tests/setup/seed');
    await seedInitialData(PB_URL);

    // Start chromedriver
    await new Promise<void>((resolve, reject) => {
      chromedriverProcess = spawn(chromedriverBin, [`--port=${CHROMEDRIVER_PORT}`], { stdio: 'pipe' });
      chromedriverProcess.on('error', reject);
      const deadline = Date.now() + 10_000;
      const poll = setInterval(async () => {
        try {
          const res = await fetch(`http://localhost:${CHROMEDRIVER_PORT}/status`);
          if (res.ok) { clearInterval(poll); resolve(); }
        } catch {
          if (Date.now() > deadline) { clearInterval(poll); reject(new Error('chromedriver did not start')); }
        }
      }, 200);
    });
  },

  onComplete() {
    if (chromedriverProcess) { chromedriverProcess.kill(); chromedriverProcess = null; }
    if (pbServer) { pbServer.kill(); pbServer = null; }
    if (tmpDir) { fs.rmSync(tmpDir, { recursive: true, force: true }); tmpDir = null; }
  },
};
