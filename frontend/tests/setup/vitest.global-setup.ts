import { spawn, execFileSync, ChildProcess } from 'child_process';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';
import { SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD } from './seed';

const PB_HOST = '127.0.0.1';
const PB_PORT = '18090';
const PB_URL = `http://${PB_HOST}:${PB_PORT}`;

let server: ChildProcess | null = null;
let tmpDir: string | null = null;

export async function setup(): Promise<void> {
  if (process.env.VITEST_UNIT_ONLY === '1') return;

  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'pb-test-'));

  const pbBin = path.resolve(
    __dirname,
    '../../../pocketbaseserver/build/pocketbaseserver',
  );

  // PocketBase 0.22+ requires existing auth to create superusers via HTTP API,
  // even on a fresh instance. Use the CLI instead: it applies migrations and
  // writes the record before the HTTP server starts.
  execFileSync(
    pbBin,
    ['superuser', 'upsert', SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD, '--dir', tmpDir],
    { stdio: 'pipe' },
  );

  server = spawn(pbBin, ['serve', '--http', `${PB_HOST}:${PB_PORT}`, '--dir', tmpDir], {
    stdio: 'pipe',
  });

  server.on('error', (err) => {
    throw new Error(`Failed to start pocketbaseserver: ${err.message}`);
  });

  await waitForUrl(`${PB_URL}/api/health`);

  process.env.TEST_PB_URL = PB_URL;

  const { seedInitialData } = await import('./seed');
  await seedInitialData(PB_URL);
}

export async function teardown(): Promise<void> {
  if (process.env.VITEST_UNIT_ONLY === '1') return;

  if (server) {
    server.kill();
    server = null;
  }
  if (tmpDir) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
    tmpDir = null;
  }
}

async function waitForUrl(url: string, timeoutMs = 15_000): Promise<void> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try {
      const res = await fetch(url);
      if (res.ok) return;
    } catch {
      // server not ready yet
    }
    await new Promise((r) => setTimeout(r, 200));
  }
  throw new Error(`Timed out waiting for ${url}`);
}
