import { describe, it, expect, vi, beforeEach } from 'vitest';

const { mockPb, mockCollection } = vi.hoisted(() => {
  const mockCollection = {
    getFirstListItem: vi.fn(),
    create: vi.fn(),
  };
  const mockPb = { collection: vi.fn().mockReturnValue(mockCollection) };
  return { mockPb, mockCollection };
});

vi.mock('@/store/pocketbase', () => ({ default: mockPb }));

import { userSettingsRepository } from '@/repositories/user-settings.repository';
import type { IUserSettings } from '@/schemas/users';

const userId = 'user-123';
const existingRecord: IUserSettings = { id: 'settings-1', user: userId, portalTheme: 'g10' };

describe('userSettingsRepository.getOrCreate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPb.collection.mockReturnValue(mockCollection);
  });

  it('returns existing record when getFirstListItem resolves', async () => {
    mockCollection.getFirstListItem.mockResolvedValue(existingRecord);
    const result = await userSettingsRepository.getOrCreate(userId);
    expect(result).toBe(existingRecord);
    expect(mockCollection.create).not.toHaveBeenCalled();
  });

  it('creates defaults when getFirstListItem throws "wasn\'t found"', async () => {
    mockCollection.getFirstListItem.mockRejectedValue(
      new Error("The requested resource wasn't found."),
    );
    const created: IUserSettings = { id: 'settings-new', user: userId, portalTheme: 'white' };
    mockCollection.create.mockResolvedValue(created);

    const result = await userSettingsRepository.getOrCreate(userId);
    expect(result).toBe(created);
    expect(mockCollection.create).toHaveBeenCalledWith({
      id: '',
      user: userId,
      portalTheme: 'white',
    });
  });

  it('rethrows other errors without calling create', async () => {
    mockCollection.getFirstListItem.mockRejectedValue(new Error('Network error'));
    await expect(userSettingsRepository.getOrCreate(userId)).rejects.toThrow('Network error');
    expect(mockCollection.create).not.toHaveBeenCalled();
  });
});
