import { describe, it, expect } from 'vitest';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

const baseUrl = process.env.TEST_PB_URL!;

async function getUserToken(): Promise<string> {
  const res = await fetch(`${baseUrl}/api/collections/users/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
  });
  const data = await res.json();
  return data.token as string;
}

describe('monthlyCountReports contract', () => {
  it('unauthenticated list returns 200 (public collection)', async () => {
    const res = await fetch(`${baseUrl}/api/collections/monthlyCountReports/records`);
    expect(res.status).toBe(200);
  });

  it('unauthenticated list returns seed records', async () => {
    const res = await fetch(`${baseUrl}/api/collections/monthlyCountReports/records`);
    const data = await res.json();
    expect(data.totalItems).toBeGreaterThan(0);
  });

  it('authenticated list also returns 200', async () => {
    const token = await getUserToken();
    const res = await fetch(`${baseUrl}/api/collections/monthlyCountReports/records`, {
      headers: { Authorization: token },
    });
    expect(res.status).toBe(200);
  });
});
