<template>
  <div>
  <CvHeader aria-label="SkillSurvey">
    <CvHeaderName prefix="">SkillSurvey</CvHeaderName>
    <template #header-global>
      <span v-if="isAuthenticated" class="bx--header__name">{{ currentUser?.email }}</span>
      <span v-else class="bx--header__name">Authentication error.</span>
      <CvButton kind="ghost" data-testid="logout-btn" @click="onLogout">Logout</CvButton>
    </template>
  </CvHeader>
  <CvSideNav :fixed="true">
    <CvSideNavItems>
      <CvSideNavLink :to="{ name: 'user-monthly-count-report' }">Monthly count report</CvSideNavLink>
      <CvSideNavLink :to="{ name: 'user-settings' }">Settings</CvSideNavLink>
    </CvSideNavItems>
  </CvSideNav>
  <CvContent>
    <router-view />
  </CvContent>
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
</script>
