import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory, RouterView } from 'vue-router';
import CarbonVue3 from '@carbon/vue';
import { routes } from '@/routes';
import pb from '@/store/pocketbase';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

vi.mock('@carbon/charts-vue', () => ({
  CcvLineChart: { template: '<svg />' },
}));

vi.mock('@/repositories/monthly-count-report.repository', () => ({
  monthlyCountReportRepository: { getAll: vi.fn().mockResolvedValue([]) },
}));

async function authenticateTestUser(): Promise<void> {
  const res = await fetch(`${process.env.TEST_PB_URL}/api/collections/users/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
  });
  const data = await res.json();
  pb.authStore.save(data.token, data.record);
}

async function pathAfterVisiting(url: string): Promise<string> {
  const router = createRouter({ history: createMemoryHistory(), routes });
  await router.push(url);
  mount(RouterView, { global: { plugins: [router, CarbonVue3] } });
  await flushPromises();
  await router.isReady();
  await flushPromises();
  return router.currentRoute.value.path;
}

describe('the old report URL', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('sends a logged-out visitor to the login page', async () => {
    expect(await pathAfterVisiting('/monthly-count-report')).toBe('/login');
  });

  it('sends a logged-in visitor to the report', async () => {
    await authenticateTestUser();
    expect(await pathAfterVisiting('/monthly-count-report')).toBe('/user/monthly-count-report');
  });
});

describe('the report route', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('sends a logged-out visitor to the login page', async () => {
    expect(await pathAfterVisiting('/user/monthly-count-report')).toBe('/login');
  });

  it('stays on the report for a logged-in visitor', async () => {
    await authenticateTestUser();
    expect(await pathAfterVisiting('/user/monthly-count-report')).toBe('/user/monthly-count-report');
  });
});
