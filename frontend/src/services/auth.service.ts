import { authRepository } from '@/repositories/auth.repository';

export const authService = {
  get isAuthenticated(): boolean {
    return authRepository.isAuthenticated;
  },

  get currentUser() {
    return authRepository.currentUser;
  },

  async login(email: string, password: string) {
    return authRepository.login(email, password);
  },

  async register(name: string, email: string, password: string, passwordConfirm: string) {
    return authRepository.register(name, email, password, passwordConfirm);
  },

  logout() {
    authRepository.logout();
  },
};
