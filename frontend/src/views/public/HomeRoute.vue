<template>
  <div>
    <h1>SkillSurvey results interface</h1>
  </div>
</template>

<script lang="ts" setup>
  import { useRouter } from 'vue-router';

  const router = useRouter();
  router.push('/login');
</script>
