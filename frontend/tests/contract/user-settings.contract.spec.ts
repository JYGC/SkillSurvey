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

describe('userSettings contract', () => {
  it('unauthenticated list returns 200 with 0 items', async () => {
    // PocketBase returns 200 with empty results when a filter rule (@request.auth.id)
    // evaluates to false — it does not return 403 for filter-based rules.
    // 403 is only returned when the rule is null (superadmin-only collection).
    const res = await fetch(`${baseUrl}/api/collections/userSettings/records`);
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.totalItems).toBe(0);
  });

  it('authenticated user can only see own records and returns 200', async () => {
    const token = await getUserToken();
    const res = await fetch(`${baseUrl}/api/collections/userSettings/records`, {
      headers: { Authorization: token },
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.totalItems).toBe(1);
  });

  it('viewing a non-existent record returns 404', async () => {
    const token = await getUserToken();
    const res = await fetch(`${baseUrl}/api/collections/userSettings/records/doesnotexist`, {
      headers: { Authorization: token },
    });
    expect(res.status).toBe(404);
  });
});
