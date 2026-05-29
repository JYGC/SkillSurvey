import { describe, it, expect } from 'vitest';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

const baseUrl = process.env.TEST_PB_URL!;

describe('auth contract', () => {
  it('unauthenticated POST to users/records returns 403 (self-registration disabled)', async () => {
    // PocketBase returns 403 when createRule is null (superadmin-only collection).
    const res = await fetch(`${baseUrl}/api/collections/users/records`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: 'New',
        email: 'new@example.com',
        password: 'Test1234!',
        passwordConfirm: 'Test1234!',
      }),
    });
    expect(res.status).toBe(403);
  });

  it('valid credentials return 200 with a token', async () => {
    const res = await fetch(`${baseUrl}/api/collections/users/auth-with-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(typeof data.token).toBe('string');
    expect(data.token.length).toBeGreaterThan(0);
  });

  it('invalid credentials return 400', async () => {
    const res = await fetch(`${baseUrl}/api/collections/users/auth-with-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ identity: SEED_USER_EMAIL, password: 'wrongpassword' }),
    });
    expect(res.status).toBe(400);
  });
});
