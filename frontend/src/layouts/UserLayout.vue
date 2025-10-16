<template>
  <p v-if="backendClient.isTokenValid">
    Logged in as: {{ JSON.stringify(backendClient.loggedInUser) }}
  </p>
  <p v-else>
    Failure to get authenticate user.
  </p>
  <CvButton @click="onLogout()">Logout</CvButton>
  <router-view />
</template>

<script lang="ts" setup>
  import { BackendClient } from '@/services/backend-client';
  import { useRouter } from 'vue-router';

  const backendClient = new BackendClient();
  const router = useRouter();

  if (!backendClient.isTokenValid) {
    router.push('/');
  }

  async function onLogout() {
    await backendClient.logoutAsync();
    router.push('/');
  }
</script>