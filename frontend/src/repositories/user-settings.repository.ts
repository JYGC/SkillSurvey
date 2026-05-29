import pb from '@/store/pocketbase';
import type { IUserSettings } from '@/schemas/users';

export const userSettingsRepository = {
  async getOrCreate(userId: string): Promise<IUserSettings> {
    try {
      return await pb.collection('userSettings').getFirstListItem<IUserSettings>(
        `user="${userId}"`,
        { fields: 'id,user,portalTheme' },
      );
    } catch (e: unknown) {
      if (!(e instanceof Error && e.message.includes("wasn't found"))) throw e;
      const defaults: IUserSettings = { id: '', user: userId, portalTheme: 'white' };
      return pb.collection('userSettings').create<IUserSettings>(defaults);
    }
  },
};
