describe('monthly count report (T42)', () => {
  it('svg chart element is present on the public route', async () => {
    // monthlyCountReports is a public collection; PublicLayout redirects
    // authenticated users away, so visit unauthenticated.
    await browser.url('/monthly-count-report');

    // Wait for the chart SVG to appear after data loads
    await $('svg').waitForExist({ timeout: 15000 });

    expect(await $('svg').isExisting()).toBe(true);
    expect(await $('p[data-testid="report-error"]').isExisting()).toBe(false);
  });
});
