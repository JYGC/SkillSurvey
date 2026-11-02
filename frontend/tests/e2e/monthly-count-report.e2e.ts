import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

async function waitForUrlContaining(urlPart: string): Promise<void> {
  await browser.waitUntil(
    async () => (await browser.getUrl()).includes(urlPart),
    { timeout: 10000, interval: 200 },
  );
}

async function logInAsTestUser(): Promise<void> {
  await browser.url('/login');
  await $('input').waitForExist({ timeout: 10000 });
  const inputs = await $$('input');
  await inputs[0].setValue(SEED_USER_EMAIL);
  await inputs[1].setValue(SEED_USER_PASSWORD);
  await $('button=Login').click();
  await waitForUrlContaining('/user/monthly-count-report');
}

describe('monthly count report (T42)', () => {
  it('the old URL sends a logged-out visitor to the login page', async () => {
    await browser.url('/monthly-count-report');

    await waitForUrlContaining('/login');

    await $('input').waitForExist({ timeout: 10000 });
    expect(await browser.getUrl()).not.toContain('/user/');
  });

  it('the old URL sends a logged-in visitor to the report, which shows a chart', async () => {
    await logInAsTestUser();

    await browser.url('/monthly-count-report');

    await waitForUrlContaining('/user/monthly-count-report');
    await $('svg').waitForExist({ timeout: 15000 });
    expect(await $('p[data-testid="report-error"]').isExisting()).toBe(false);
  });
});
