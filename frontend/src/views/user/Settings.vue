<template>
  {{ JSON.stringify(userSetting) }}
</template>
<script lang="ts" setup>
  import { IUserSettings } from '@/schemas/users';
  import { getBackendClient } from '@/services/backend-client';
  import { ref } from 'vue';

  const backendClient = getBackendClient();

  const userSetting = ref<IUserSettings | null>(null);

  const getUserSettings = async () => {
    if (backendClient.authStore.record == null) {
      return;
    }
    const userId = backendClient.authStore.record.id;
    try {
      userSetting.value = await backendClient.collection('user_settings').getFirstListItem<IUserSettings | null>(
        `user_settings.user="${userId}"`,
        {
          fields: 'user,portalTheme',
        }
      );
    } catch (error) {
      if (!(error instanceof Error && error.message.includes('requested resource wasn\'t found'))) {
        throw error;
      }
      userSetting.value = {
        user: userId,
        portalThemes: 'white',
      };
      await backendClient.collection('user_settings').create(userSetting.value);
    }
  };

  getUserSettings();
</script>