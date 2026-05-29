import { ref } from 'vue';
import type { IUserSettings } from '@/schemas/users';
import { authRepository } from '@/repositories/auth.repository';
import { userSettingsRepository } from '@/repositories/user-settings.repository';

export function useUserSettings() {
  const userSetting = ref<IUserSettings | null>(null);

  async function load() {
    const user = authRepository.currentUser;
    if (!user) return;
    userSetting.value = await userSettingsRepository.getOrCreate(user.id);
  }

  return { userSetting, load };
}
