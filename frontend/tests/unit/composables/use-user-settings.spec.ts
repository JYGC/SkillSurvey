import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { IUserSettings } from '@/schemas/users';

const mockGetOrCreateUserSettingsForCurrentUser = vi.hoisted(() => vi.fn());

vi.mock('@/services/user-settings.service', () => ({
  getOrCreateUserSettingsForCurrentUser: mockGetOrCreateUserSettingsForCurrentUser,
}));

const seedSettings: IUserSettings = { id: 'set1', user: 'user1', portalTheme: 'white' };

describe('useUserSettings', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
  });

  it('load() with no current user leaves userSetting null', async () => {
    mockGetOrCreateUserSettingsForCurrentUser.mockResolvedValue(null);
    const { useUserSettings } = await import('@/composables/use-user-settings');
    const { userSetting, load } = useUserSettings();
    await load();
    expect(userSetting.value).toBeNull();
  });

  it('load() with current user sets userSetting from the service', async () => {
    mockGetOrCreateUserSettingsForCurrentUser.mockResolvedValue(seedSettings);
    const { useUserSettings } = await import('@/composables/use-user-settings');
    const { userSetting, load } = useUserSettings();
    await load();
    expect(mockGetOrCreateUserSettingsForCurrentUser).toHaveBeenCalledOnce();
    expect(userSetting.value).toEqual(seedSettings);
  });
});
