<template>
  <CvFluidForm>
    <CvTextInput label="Email" v-model="email" />
    <CvTextInput label="Password" type="password" v-model="password" />
  </CvFluidForm>
  <br />
  <CvButton @click="onSubmit()">Login</CvButton>
</template>
<script lang="ts" setup>
  import { getBackendClient } from '@/services/backend-client';
  import { ref } from 'vue';
  import { useRouter } from 'vue-router';

  const backendClient = getBackendClient();
  const router = useRouter();

  const email = ref('');
  const password = ref('');

  console.log(process.env.VUE_APP_POCKETBASE_URL);

  async function login() {
    try {
      const cookieString = await backendClient.collection('users').authWithPassword(email.value, password.value);
      console.log(cookieString);
      console.log(backendClient.authStore.isValid);
      console.log(backendClient.authStore.token);
      console.log(backendClient.authStore.record);

      router.push('/user/profile'); // Redirect to user layout or dashboard after login
    } catch (error) {
      console.error('Login failed:', error);
      alert(error);
    }
  }

  function onSubmit() {
    login();
  }
</script>