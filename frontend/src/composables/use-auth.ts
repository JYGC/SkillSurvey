import { computed } from 'vue';
import { authService } from '@/services/auth.service';

export function useAuth() {
  const isAuthenticated = computed(() => authService.isAuthenticated);
  const currentUser = computed(() => authService.currentUser);

  async function login(email: string, password: string) {
    return authService.login(email, password);
  }

  async function register(name: string, email: string, password: string, passwordConfirm: string) {
    return authService.register(name, email, password, passwordConfirm);
  }

  function logout() {
    authService.logout();
  }

  return { isAuthenticated, currentUser, login, register, logout };
}
