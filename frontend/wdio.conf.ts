import type { Options } from '@wdio/types';
import { spawn, ChildProcess } from 'child_process';

const baseUrl = process.env.TEST_E2E_URL ?? 'http://localhost:8080';
const chromeBin = process.env.CHROMIUM_PATH ?? '/usr/local/bin/chrome';
const chromedriverBin = process.env.CHROMEDRIVER_PATH ?? '/usr/local/bin/chromedriver';
const CHROMEDRIVER_PORT = 9515;

let chromedriverProcess: ChildProcess | null = null;

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
    },
  ],
  baseUrl,

  async onPrepare() {
    await new Promise<void>((resolve, reject) => {
      chromedriverProcess = spawn(chromedriverBin, [`--port=${CHROMEDRIVER_PORT}`], { stdio: 'pipe' });
      chromedriverProcess.on('error', reject);
      // Poll until chromedriver is ready
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
  },
};
