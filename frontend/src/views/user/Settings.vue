<template>
  <div>{{ JSON.stringify(userSetting) }}</div>
</template>
<script lang="ts" setup>
import { onMounted } from 'vue';
import { useUserSettings } from '@/composables/use-user-settings';

const { userSetting, load } = useUserSettings();

onMounted(load);
</script>
