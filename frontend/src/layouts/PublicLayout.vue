<template>
  <div class="row">
    <div class="col-md-2 nav-div">
      <b-nav vertical>
        <b-nav-item to="/" active>Home</b-nav-item>
        <b-nav-item to="/monthly-count-report">Monthly count report</b-nav-item>
        <b-nav-item class="expand">
          <b-button v-b-toggle.collapse-skill-details variant="primary" class="w-100">Skill details</b-button>
          <b-collapse id="collapse-skill-details">
            <b-nav-item to="/skill-list">Skills list</b-nav-item>
            <b-nav-item to="/skill-type-list">Skill type list</b-nav-item>
          </b-collapse>
        </b-nav-item>
      </b-nav>
    </div>
    <div class="col-md-10">
      <router-view />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { getBackendClient } from '@/services/backend-client';
import { useRouter } from 'vue-router';

  const backendClient = getBackendClient();
  const router = useRouter();

  if (backendClient.authStore.isValid) {
    router.push('/user/profile');
  }
</script>