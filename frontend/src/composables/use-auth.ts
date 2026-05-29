import { computed } from 'vue';
import { authRepository } from '@/repositories/auth.repository';

export function useAuth() {
  const isAuthenticated = computed(() => authRepository.isAuthenticated);
  const currentUser = computed(() => authRepository.currentUser);

  async function login(email: string, password: string) {
    return authRepository.login(email, password);
  }

  function logout() {
    authRepository.logout();
  }

  return { isAuthenticated, currentUser, login, logout };
}
