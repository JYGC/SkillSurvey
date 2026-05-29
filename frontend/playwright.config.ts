import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  use: {
    baseURL: process.env.TEST_E2E_URL ?? 'http://localhost:8080',
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        executablePath: process.env.CHROMIUM_PATH,
        launchOptions: { args: ['--no-sandbox'] },
      },
    },
  ],
});
