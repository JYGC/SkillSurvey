<template>
  <div>
    <p v-if="isAuthenticated">
      Logged in as: {{ currentUser?.email }}
    </p>
    <p v-else>
      Failure to get authenticate user.
    </p>
    <CvButton data-testid="logout-btn" @click="onLogout()">Logout</CvButton>
    <CvButton @click="onSettingsClick()">Settings</CvButton>
    <router-view />
  </div>
</template>

<script lang="ts" setup>
import { useAuth } from '@/composables/use-auth';
import { useRouter } from 'vue-router';

const { isAuthenticated, currentUser, logout } = useAuth();
const router = useRouter();

if (!isAuthenticated.value) {
  router.push('/');
}

const onLogout = () => {
  logout();
  router.push('/');
};

const onSettingsClick = () => {
  router.push('/user/settings');
};
</script>
