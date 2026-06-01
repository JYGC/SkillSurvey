import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockAuthRepository = vi.hoisted(() => ({
  isAuthenticated: false,
  currentUser: null as { id: string; email: string } | null,
  login: vi.fn(),
  logout: vi.fn(),
}));

vi.mock('@/repositories/auth.repository', () => ({
  authRepository: mockAuthRepository,
}));

describe('useAuth', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockAuthRepository.isAuthenticated = false;
    mockAuthRepository.currentUser = null;
  });

  it('isAuthenticated mirrors authRepository.isAuthenticated when false', async () => {
    mockAuthRepository.isAuthenticated = false;
    const { useAuth } = await import('@/composables/use-auth');
    const { isAuthenticated } = useAuth();
    expect(isAuthenticated.value).toBe(false);
  });

  it('isAuthenticated mirrors authRepository.isAuthenticated when true', async () => {
    mockAuthRepository.isAuthenticated = true;
    const { useAuth } = await import('@/composables/use-auth');
    const { isAuthenticated } = useAuth();
    expect(isAuthenticated.value).toBe(true);
  });

  it('currentUser mirrors authRepository.currentUser', async () => {
    const user = { id: 'user1', email: 'user@example.com' };
    mockAuthRepository.currentUser = user;
    const { useAuth } = await import('@/composables/use-auth');
    const { currentUser } = useAuth();
    expect(currentUser.value).toBe(user);
  });

  it('login delegates to authRepository.login', async () => {
    mockAuthRepository.login.mockResolvedValue({ token: 'tok' });
    const { useAuth } = await import('@/composables/use-auth');
    const { login } = useAuth();
    await login('a@b.com', 'pass');
    expect(mockAuthRepository.login).toHaveBeenCalledWith('a@b.com', 'pass');
  });

  it('login propagates rejection from authRepository.login', async () => {
    mockAuthRepository.login.mockRejectedValue(new Error('bad credentials'));
    const { useAuth } = await import('@/composables/use-auth');
    const { login } = useAuth();
    await expect(login('a@b.com', 'wrong')).rejects.toThrow('bad credentials');
  });

  it('logout delegates to authRepository.logout', async () => {
    const { useAuth } = await import('@/composables/use-auth');
    const { logout } = useAuth();
    logout();
    expect(mockAuthRepository.logout).toHaveBeenCalledOnce();
  });
});
