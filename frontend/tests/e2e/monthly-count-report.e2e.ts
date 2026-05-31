describe('monthly count report (T42)', () => {
  it('svg chart element is present on the public route', async () => {
    await browser.url('/monthly-count-report');

    await $('svg').waitForExist({ timeout: 15000 });

    expect(await $('svg').isExisting()).toBe(true);
    expect(await $('p[data-testid="report-error"]').isExisting()).toBe(false);
  });
});
