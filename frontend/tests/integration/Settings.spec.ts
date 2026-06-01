import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import Settings from '@/views/user/Settings.vue';
import type { IUserSettings } from '@/schemas/users';

const mockGetOrCreate = vi.hoisted(() => vi.fn());
const mockCurrentUser = vi.hoisted(() => ({ value: { id: 'user1' } as { id: string } | null }));

vi.mock('@/repositories/user-settings.repository', () => ({
  userSettingsRepository: { getOrCreate: mockGetOrCreate },
}));

vi.mock('@/repositories/auth.repository', () => ({
  authRepository: {
    get currentUser() { return mockCurrentUser.value; },
  },
}));

describe('Settings', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockCurrentUser.value = { id: 'user1' };
  });

  it('authenticated user: portalTheme visible in rendered output', async () => {
    const seedSettings: IUserSettings = { id: 'set1', user: 'user1', portalTheme: 'white' };
    mockGetOrCreate.mockResolvedValue(seedSettings);

    const wrapper = mount(Settings);
    await flushPromises();

    expect(mockGetOrCreate).toHaveBeenCalledWith('user1');
    expect(wrapper.text()).toContain('white');
  });
});
