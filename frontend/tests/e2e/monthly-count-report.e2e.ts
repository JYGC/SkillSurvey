describe('monthly count report (T42)', () => {
  it('svg chart element is present on the public route', async () => {
    await browser.url('/monthly-count-report');

    // Wait for Vue to mount and data to load
    await browser.pause(5000);

    const html = (await browser.getPageSource()).substring(0, 4000);
    console.log('--- PAGE SOURCE ---\n', html);

    await $('svg').waitForExist({ timeout: 15000 });

    expect(await $('svg').isExisting()).toBe(true);
    expect(await $('p[data-testid="report-error"]').isExisting()).toBe(false);
  });
});
