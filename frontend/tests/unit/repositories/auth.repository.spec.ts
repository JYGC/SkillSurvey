import { describe, it, expect, vi, beforeEach } from 'vitest';

const { mockPb, mockCollection } = vi.hoisted(() => {
  const mockCollection = {
    authWithPassword: vi.fn().mockResolvedValue({}),
    create: vi.fn().mockResolvedValue({}),
  };
  const mockPb = {
    authStore: {
      isValid: false as boolean,
      record: null as unknown,
      clear: vi.fn(),
    },
    collection: vi.fn().mockReturnValue(mockCollection),
  };
  return { mockPb, mockCollection };
});

vi.mock('@/store/pocketbase', () => ({ default: mockPb }));

import { authRepository } from '@/repositories/auth.repository';

describe('authRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPb.authStore.isValid = false;
    mockPb.authStore.record = null;
    mockPb.collection.mockReturnValue(mockCollection);
  });

  it('isAuthenticated reflects pb.authStore.isValid', () => {
    mockPb.authStore.isValid = true;
    expect(authRepository.isAuthenticated).toBe(true);
    mockPb.authStore.isValid = false;
    expect(authRepository.isAuthenticated).toBe(false);
  });

  it('currentUser reflects pb.authStore.record', () => {
    const user = { id: 'u1', email: 'test@test.com' };
    mockPb.authStore.record = user;
    expect(authRepository.currentUser).toBe(user);
  });

  it('login delegates to pb.collection(users).authWithPassword', async () => {
    await authRepository.login('a@b.com', 'pass');
    expect(mockPb.collection).toHaveBeenCalledWith('users');
    expect(mockCollection.authWithPassword).toHaveBeenCalledWith('a@b.com', 'pass');
  });

  it('register delegates to pb.collection(users).create with all fields', async () => {
    await authRepository.register('Test', 'a@b.com', 'pass', 'pass');
    expect(mockPb.collection).toHaveBeenCalledWith('users');
    expect(mockCollection.create).toHaveBeenCalledWith({
      name: 'Test',
      email: 'a@b.com',
      password: 'pass',
      passwordConfirm: 'pass',
    });
  });

  it('logout delegates to pb.authStore.clear', () => {
    authRepository.logout();
    expect(mockPb.authStore.clear).toHaveBeenCalled();
  });
});
