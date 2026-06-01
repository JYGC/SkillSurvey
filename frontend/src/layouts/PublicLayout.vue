<template>
  <div class="row">
    <div class="col-md-2 nav-div">
      <b-nav vertical>
        <b-nav-item to="/" active>Home</b-nav-item>
        <b-nav-item to="/monthly-count-report">Monthly count report</b-nav-item>
      </b-nav>
    </div>
    <div class="col-md-10">
      <router-view />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useAuth } from '@/composables/use-auth';
import { useRouter } from 'vue-router';

const { isAuthenticated } = useAuth();
const router = useRouter();

if (isAuthenticated.value) {
  router.push('/user/profile');
}
</script>
