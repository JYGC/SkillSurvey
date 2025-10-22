<template>
  <CvFluidForm>
    <CvTextInput label="Email" v-model="email" />
    <CvTextInput label="Password" type="password" v-model="password" />
  </CvFluidForm>
  <br />
  <CvButton @click="onSubmit()">Login</CvButton>
  <br />
  <br />
  <br />
  <p>Don't have an account?</p>
  <CvLink href="/register">Register</CvLink>
</template>
<script lang="ts" setup>
  import { getBackendClient } from '@/services/backend-client';
  import { ref } from 'vue';
  import { useRouter } from 'vue-router';

  const backendClient = getBackendClient();
  const router = useRouter();

  const email = ref('');
  const password = ref('');

  const login = () => {
    try {
      backendClient.collection('users').authWithPassword(email.value, password.value);
      router.push('/user/profile'); // Redirect to user layout or dashboard after login
    } catch (error) {
      console.error('Login failed:', error);
      alert(error);
    }
  }

  const onSubmit = () => {
    login();
  }
</script>