import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockAuthService = vi.hoisted(() => ({
  isAuthenticated: false,
  currentUser: null as { id: string; email: string } | null,
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
}));

vi.mock('@/services/auth.service', () => ({
  authService: mockAuthService,
}));

describe('useAuth', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockAuthService.isAuthenticated = false;
    mockAuthService.currentUser = null;
  });

  it('isAuthenticated mirrors authService.isAuthenticated when false', async () => {
    mockAuthService.isAuthenticated = false;
    const { useAuth } = await import('@/composables/use-auth');
    const { isAuthenticated } = useAuth();
    expect(isAuthenticated.value).toBe(false);
  });

  it('isAuthenticated mirrors authService.isAuthenticated when true', async () => {
    mockAuthService.isAuthenticated = true;
    const { useAuth } = await import('@/composables/use-auth');
    const { isAuthenticated } = useAuth();
    expect(isAuthenticated.value).toBe(true);
  });

  it('currentUser mirrors authService.currentUser', async () => {
    const user = { id: 'user1', email: 'user@example.com' };
    mockAuthService.currentUser = user;
    const { useAuth } = await import('@/composables/use-auth');
    const { currentUser } = useAuth();
    expect(currentUser.value).toBe(user);
  });

  it('login delegates to authService.login', async () => {
    mockAuthService.login.mockResolvedValue({ token: 'tok' });
    const { useAuth } = await import('@/composables/use-auth');
    const { login } = useAuth();
    await login('a@b.com', 'pass');
    expect(mockAuthService.login).toHaveBeenCalledWith('a@b.com', 'pass');
  });

  it('login propagates rejection from authService.login', async () => {
    mockAuthService.login.mockRejectedValue(new Error('bad credentials'));
    const { useAuth } = await import('@/composables/use-auth');
    const { login } = useAuth();
    await expect(login('a@b.com', 'wrong')).rejects.toThrow('bad credentials');
  });

  it('register delegates to authService.register', async () => {
    mockAuthService.register.mockResolvedValue({ id: 'user1' });
    const { useAuth } = await import('@/composables/use-auth');
    const { register } = useAuth();
    await register('Test', 'a@b.com', 'pass', 'pass');
    expect(mockAuthService.register).toHaveBeenCalledWith('Test', 'a@b.com', 'pass', 'pass');
  });

  it('logout delegates to authService.logout', async () => {
    const { useAuth } = await import('@/composables/use-auth');
    const { logout } = useAuth();
    logout();
    expect(mockAuthService.logout).toHaveBeenCalledOnce();
  });
});
