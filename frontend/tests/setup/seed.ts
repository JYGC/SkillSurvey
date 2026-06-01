// Credentials written to process.env for test files to consume.
export const SEED_ADMIN_EMAIL = 'test-admin@skillsurvey-test.internal';
export const SEED_ADMIN_PASSWORD = 'TestAdmin1234!';
export const SEED_USER_EMAIL = 'test-user@skillsurvey-test.internal';
export const SEED_USER_PASSWORD = 'TestUser1234!';

// Seed months: 3 consecutive months of data for one skill.
const SEED_MONTHS = [
  { YearMonth: '2024-10', yearMonthDate: '2024-10-01 00:00:00.000Z', count: 10 },
  { YearMonth: '2024-11', yearMonthDate: '2024-11-01 00:00:00.000Z', count: 15 },
  { YearMonth: '2024-12', yearMonthDate: '2024-12-01 00:00:00.000Z', count: 20 },
];

export async function seedInitialData(baseUrl: string): Promise<void> {
  // The superadmin is created via `pocketbaseserver superuser upsert` in
  // vitest.global-setup.ts before the server starts. Authenticate here.
  const authRes = await post(baseUrl, '/api/collections/_superusers/auth-with-password', undefined, {
    identity: SEED_ADMIN_EMAIL,
    password: SEED_ADMIN_PASSWORD,
  });
  const adminToken: string = authRes.token;

  // Create test user (self-registration is disabled; superadmin creates accounts).
  const user = await post(baseUrl, '/api/collections/users/records', adminToken, {
    name: 'Test User',
    email: SEED_USER_EMAIL,
    password: SEED_USER_PASSWORD,
    passwordConfirm: SEED_USER_PASSWORD,
  });

  // Create userSettings for the test user.
  await post(baseUrl, '/api/collections/userSettings/records', adminToken, {
    user: user.id,
    portalTheme: 'white',
  });

  // Create a skillType for test monthlyCountReports.
  const skillType = await post(baseUrl, '/api/collections/skillTypes/records', adminToken, {
    name: 'Test Skill Type',
    description: 'Used by automated tests',
  });

  // Create a skillName under that type.
  const skillName = await post(baseUrl, '/api/collections/skillNames/records', adminToken, {
    name: 'TestSkill',
    isEnabled: true,
    skillType: skillType.id,
  });

  // Create seed monthlyCountReports.
  for (const month of SEED_MONTHS) {
    await post(baseUrl, '/api/collections/monthlyCountReports/records', adminToken, {
      identifier: `TestSkill-${month.YearMonth}`,
      YearMonth: month.YearMonth,
      yearMonthDate: month.yearMonthDate,
      count: month.count,
      skillName: skillName.id,
    });
  }

  process.env.TEST_USER_EMAIL = SEED_USER_EMAIL;
  process.env.TEST_USER_PASSWORD = SEED_USER_PASSWORD;
}

async function post(baseUrl: string, path: string, token: string | undefined, body: unknown): Promise<any> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = token;
  const res = await fetch(`${baseUrl}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`POST ${path} → HTTP ${res.status}: ${text}`);
  }
  return res.json();
}
