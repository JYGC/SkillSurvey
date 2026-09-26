import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockAuthRepository = vi.hoisted(() => ({
  isAuthenticated: false,
  currentUser: null as { id: string; email: string } | null,
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
}));

vi.mock('@/repositories/auth.repository', () => ({
  authRepository: mockAuthRepository,
}));

describe('authService', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockAuthRepository.isAuthenticated = false;
    mockAuthRepository.currentUser = null;
  });

  it('isAuthenticated mirrors authRepository.isAuthenticated', async () => {
    mockAuthRepository.isAuthenticated = true;
    const { authService } = await import('@/services/auth.service');
    expect(authService.isAuthenticated).toBe(true);
  });

  it('currentUser mirrors authRepository.currentUser', async () => {
    const user = { id: 'user1', email: 'user@example.com' };
    mockAuthRepository.currentUser = user;
    const { authService } = await import('@/services/auth.service');
    expect(authService.currentUser).toBe(user);
  });

  it('login delegates to authRepository.login', async () => {
    mockAuthRepository.login.mockResolvedValue({ token: 'tok' });
    const { authService } = await import('@/services/auth.service');
    await authService.login('a@b.com', 'pass');
    expect(mockAuthRepository.login).toHaveBeenCalledWith('a@b.com', 'pass');
  });

  it('register delegates to authRepository.register', async () => {
    mockAuthRepository.register.mockResolvedValue({ id: 'user1' });
    const { authService } = await import('@/services/auth.service');
    await authService.register('Test', 'a@b.com', 'pass', 'pass');
    expect(mockAuthRepository.register).toHaveBeenCalledWith('Test', 'a@b.com', 'pass', 'pass');
  });

  it('logout delegates to authRepository.logout', async () => {
    const { authService } = await import('@/services/auth.service');
    authService.logout();
    expect(mockAuthRepository.logout).toHaveBeenCalledOnce();
  });
});
