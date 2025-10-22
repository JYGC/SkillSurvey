<template>
  {{ JSON.stringify(userSettings) }}
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
    userSetting.value = await backendClient.collection('user_settings').getOne<IUserSettings | null>(userId, {
      fields: 'user,portalTheme',
    });
    if (userSetting.value == null) {
      userSetting.value = {
        user: userId,
        portalThemes: 'white',
      };
      await backendClient.collection('user_settings').create(userSetting.value);
    }

  };

  getUserSettings();
</script>