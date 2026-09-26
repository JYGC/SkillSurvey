import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { IUserSettings } from '@/schemas/users';

const mockAuthService = vi.hoisted(() => ({
  currentUser: null as { id: string } | null,
}));

const mockUserSettingsRepository = vi.hoisted(() => ({
  getOrCreate: vi.fn(),
}));

vi.mock('@/services/auth.service', () => ({
  authService: mockAuthService,
}));

vi.mock('@/repositories/user-settings.repository', () => ({
  userSettingsRepository: mockUserSettingsRepository,
}));

const seedSettings: IUserSettings = { id: 'set1', user: 'user1', portalTheme: 'white' };

describe('getOrCreateUserSettingsForCurrentUser', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
    mockAuthService.currentUser = null;
  });

  it('returns null and does not call the repository when there is no current user', async () => {
    mockAuthService.currentUser = null;
    const { getOrCreateUserSettingsForCurrentUser } = await import('@/services/user-settings.service');
    const result = await getOrCreateUserSettingsForCurrentUser();
    expect(mockUserSettingsRepository.getOrCreate).not.toHaveBeenCalled();
    expect(result).toBeNull();
  });

  it('delegates to the repository with the current user id', async () => {
    mockAuthService.currentUser = { id: 'user1' };
    mockUserSettingsRepository.getOrCreate.mockResolvedValue(seedSettings);
    const { getOrCreateUserSettingsForCurrentUser } = await import('@/services/user-settings.service');
    const result = await getOrCreateUserSettingsForCurrentUser();
    expect(mockUserSettingsRepository.getOrCreate).toHaveBeenCalledWith('user1');
    expect(result).toEqual(seedSettings);
  });
});
