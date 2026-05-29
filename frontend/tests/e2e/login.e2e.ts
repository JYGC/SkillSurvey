import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

describe('login and logout flow', () => {
  it('valid credentials navigate to /user/profile (T41)', async () => {
    await browser.url('/login');

    // Wait for Vue to render the login form before collecting inputs
    await $('input').waitForExist({ timeout: 10000 });
    const inputs = await $$('input');
    await inputs[0].setValue(SEED_USER_EMAIL);
    await inputs[1].setValue(SEED_USER_PASSWORD);

    await $('button=Login').click();

    await browser.waitUntil(
      async () => (await browser.getUrl()).includes('/user/profile'),
      { timeout: 10000, interval: 200 },
    );

    expect(await browser.getUrl()).toContain('/user/profile');
  });

  it('logout redirects to / (T43)', async () => {
    // Continues from the logged-in state of the previous test
    await $('[data-testid="logout-btn"]').click();

    await browser.waitUntil(
      async () => !(await browser.getUrl()).includes('/user/profile'),
      { timeout: 10000, interval: 200 },
    );

    const finalUrl = await browser.getUrl();
    expect(finalUrl).not.toContain('/user/');
  });
});
