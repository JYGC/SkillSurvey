import pb from '@/store/pocketbase';

export const authRepository = {
  get isAuthenticated(): boolean {
    return pb.authStore.isValid;
  },

  get currentUser() {
    return pb.authStore.record;
  },

  async login(email: string, password: string) {
    return pb.collection('users').authWithPassword(email, password);
  },

  async register(name: string, email: string, password: string, passwordConfirm: string) {
    return pb.collection('users').create({ name, email, password, passwordConfirm });
  },

  logout() {
    pb.authStore.clear();
  },
};
