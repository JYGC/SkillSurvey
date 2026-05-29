import { describe, it, expect, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createRouter, createMemoryHistory } from 'vue-router';
import UserLayout from '@/layouts/UserLayout.vue';
import pb from '@/store/pocketbase';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

function makeRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div/>' } },
      { path: '/user', component: UserLayout, children: [{ path: 'profile', component: { template: '<div/>' } }] },
      { path: '/user/settings', component: { template: '<div/>' } },
    ],
  });
}

const stubs = { CvButton: true, RouterView: true };

async function authenticateTestUser(): Promise<void> {
  const res = await fetch(`${process.env.TEST_PB_URL}/api/collections/users/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
  });
  const data = await res.json();
  pb.authStore.save(data.token, data.record);
}

describe('UserLayout', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('unauthenticated: redirects to /', async () => {
    const router = makeRouter();
    await router.push('/user/profile');
    mount(UserLayout, { global: { plugins: [router], stubs } });
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/');
  });

  it('authenticated: renders user email and Logout button', async () => {
    await authenticateTestUser();
    const router = makeRouter();
    await router.push('/user/profile');
    const wrapper = mount(UserLayout, { global: { plugins: [router], stubs } });
    await flushPromises();
    expect(router.currentRoute.value.path).toBe('/user/profile');
    expect(wrapper.text()).toContain(SEED_USER_EMAIL);
  });

  it('Logout clears auth and redirects to /', async () => {
    await authenticateTestUser();
    const router = makeRouter();
    await router.push('/user/profile');
    const wrapper = mount(UserLayout, { global: { plugins: [router], stubs } });
    await flushPromises();
    await wrapper.find('[data-testid="logout-btn"]').trigger('click');
    await flushPromises();
    expect(pb.authStore.isValid).toBe(false);
    expect(router.currentRoute.value.path).toBe('/');
  });
});
