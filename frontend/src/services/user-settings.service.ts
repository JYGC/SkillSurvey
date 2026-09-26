import type { IUserSettings } from '@/schemas/users';
import { authService } from '@/services/auth.service';
import { userSettingsRepository } from '@/repositories/user-settings.repository';

export async function getOrCreateUserSettingsForCurrentUser(): Promise<IUserSettings | null> {
  const user = authService.currentUser;
  if (!user) return null;
  return userSettingsRepository.getOrCreate(user.id);
}
