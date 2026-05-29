import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { IUserSettings } from '@/schemas/users';

const mockAuthRepository = vi.hoisted(() => ({
  currentUser: null as { id: string } | null,
}));

const mockUserSettingsRepository = vi.hoisted(() => ({
  getOrCreate: vi.fn(),
}));

vi.mock('@/repositories/auth.repository', () => ({
  authRepository: mockAuthRepository,
}));

vi.mock('@/repositories/user-settings.repository', () => ({
  userSettingsRepository: mockUserSettingsRepository,
}));

const seedSettings: IUserSettings = { id: 'set1', user: 'user1', portalTheme: 'white' };

describe('useUserSettings', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
    mockAuthRepository.currentUser = null;
  });

  it('load() with no current user leaves userSetting null and does not call repository', async () => {
    mockAuthRepository.currentUser = null;
    const { useUserSettings } = await import('@/composables/use-user-settings');
    const { userSetting, load } = useUserSettings();
    await load();
    expect(mockUserSettingsRepository.getOrCreate).not.toHaveBeenCalled();
    expect(userSetting.value).toBeNull();
  });

  it('load() with current user calls repository and sets userSetting', async () => {
    mockAuthRepository.currentUser = { id: 'user1' };
    mockUserSettingsRepository.getOrCreate.mockResolvedValue(seedSettings);
    const { useUserSettings } = await import('@/composables/use-user-settings');
    const { userSetting, load } = useUserSettings();
    await load();
    expect(mockUserSettingsRepository.getOrCreate).toHaveBeenCalledWith('user1');
    expect(userSetting.value).toEqual(seedSettings);
  });
});
