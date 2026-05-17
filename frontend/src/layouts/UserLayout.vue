<template>
  <p v-if="backendClient.authStore.isValid">
    Logged in as: {{ JSON.stringify(backendClient.authStore.record) }}
  </p>
  <p v-else>
    Failure to get authenticate user.
  </p>
  <CvButton @click="onSettingsClick()">Settings</CvButton>
  <CvButton @click="onLogout()">Logout</CvButton>
  <router-view />
</template>

<script lang="ts" setup>
  import { getBackendClient } from '@/services/backend-client';
  import { useRouter } from 'vue-router';

  const backendClient = getBackendClient();
  const router = useRouter();

  if (!backendClient.authStore.isValid) {
    router.push('/');
  }

  const onLogout = () => {
    backendClient.authStore.clear();
    router.push('/');
  }

  const onSettingsClick = () => {
    router.push('/user/settings');
  }
</script>