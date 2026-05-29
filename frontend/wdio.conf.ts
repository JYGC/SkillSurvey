import type { Options } from '@wdio/types';

const baseUrl = process.env.TEST_E2E_URL ?? 'http://localhost:8080';
const chromeBin = process.env.CHROMIUM_PATH ?? '/usr/local/bin/chrome';

export const config: Options.Testrunner = {
  runner: 'local',
  specs: ['./tests/e2e/**/*.e2e.ts'],
  framework: 'mocha',
  reporters: ['spec'],
  mochaOpts: { timeout: 30000 },
  capabilities: [
    {
      browserName: 'chrome',
      'goog:chromeOptions': {
        binary: chromeBin,
        args: ['--headless', '--no-sandbox', '--disable-dev-shm-usage'],
      },
    },
  ],
  services: [
    [
      'chromedriver',
      {
        chromedriverCustomPath: process.env.CHROMEDRIVER_PATH ?? '/usr/local/bin/chromedriver',
      },
    ],
  ],
  baseUrl,
};
