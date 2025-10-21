<template>
  {{ JSON.stringify(userSettings) }}
</template>
<script lang="ts" setup>
  import { IUserSettings } from '@/schemas/users';
  import { getBackendClient } from '@/services/backend-client';
  import { ref } from 'vue';

  const backendClient = getBackendClient();

  const userSettings = ref<IUserSettings | null>(null);

  const getUserSettings = async () => {
    if (backendClient.authStore.record == null) {
      return;
    }
    const userId = backendClient.authStore.record.id;
    const userSetting = backendClient.collection('user_settings').getOne(userId, {
      fields: 'user,portalTheme',
    });
    if (userSetting == null) {
      userSettings.value = {
        user: userId,
        portalThemes: 'white',
      };
    }

  };

  getUserSettings();
</script>