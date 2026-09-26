import { ref } from 'vue';
import type { IUserSettings } from '@/schemas/users';
import { getOrCreateUserSettingsForCurrentUser } from '@/services/user-settings.service';

export function useUserSettings() {
  const userSetting = ref<IUserSettings | null>(null);

  async function load() {
    userSetting.value = await getOrCreateUserSettingsForCurrentUser();
  }

  return { userSetting, load };
}
