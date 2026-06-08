<template>
  <div class="public-layout">
    <router-view />
  </div>
</template>

<script lang="ts" setup>
import { useAuth } from '@/composables/use-auth';
import { useRouter } from 'vue-router';

const { isAuthenticated } = useAuth();
const router = useRouter();

if (isAuthenticated.value) {
  router.push('/user/monthly-count-report');
}
</script>
