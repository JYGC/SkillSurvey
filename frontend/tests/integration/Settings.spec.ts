import { describe, it, expect, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import Settings from '@/views/user/Settings.vue';
import pb from '@/store/pocketbase';
import { SEED_USER_EMAIL, SEED_USER_PASSWORD } from '../setup/seed';

async function authenticateTestUser(): Promise<void> {
  const res = await fetch(`${process.env.TEST_PB_URL}/api/collections/users/auth-with-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ identity: SEED_USER_EMAIL, password: SEED_USER_PASSWORD }),
  });
  const data = await res.json();
  pb.authStore.save(data.token, data.record);
}

describe('Settings', () => {
  beforeEach(() => {
    pb.authStore.clear();
  });

  it('authenticated user: portalTheme visible in rendered output', async () => {
    await authenticateTestUser();
    const wrapper = mount(Settings);
    await flushPromises();
    expect(wrapper.text()).toContain('white');
  });
});
