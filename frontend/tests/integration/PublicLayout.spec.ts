import { describe, it, expect, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import PublicLayout from '@/layouts/PublicLayout.vue';
import pb from '@/store/pocketbase';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: PublicLayout, children: [{ path: '', component: { template: '<div/>' } }] },
      { path: '/user/profile', component: { template: '<div/>' } },
    ],
  });
}

const stubs = {
  'b-nav': true,
  'b-nav-item': true,
  'b-button': true,
  'b-collapse': true,
  RouterView: true,
};

async function authenticateTestUser(): Promise<void> {
  const res = await fetch(`${process.env.TEST_PB_URL}/api/collections/users/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
  });
  const data = await res.json();
  pb.authStore.save(data.token, data.record);
}

describe('PublicLayout', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('unauthenticated: no redirect, component renders', async () => {
    const router = makeRouter();
    await router.push('/');
    mount(PublicLayout, { global: { plugins: [router], stubs } });
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/');
  });

  it('authenticated: redirects to /user/profile', async () => {
    await authenticateTestUser();
    const router = makeRouter();
    await router.push('/');
    mount(PublicLayout, { global: { plugins: [router], stubs } });
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/user/profile');
  });
});
