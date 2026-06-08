<template>
  <div>
    <CvFluidForm>
      <CvTextInput label="Email" v-model="email" />
      <CvTextInput label="Password" type="password" v-model="password" />
    </CvFluidForm>
    <br />
    <CvButton @click="onSubmit()">Login</CvButton>
    <p v-if="loginError" data-testid="login-error">{{ loginError }}</p>
    <br />
    <br />
    <p>Don't have an account?</p>
    <CvLink href="/register">Register</CvLink>
  </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuth } from '@/composables/use-auth';

const router = useRouter();
const { login } = useAuth();

const email = ref('');
const password = ref('');
const loginError = ref('');

const onSubmit = async () => {
  loginError.value = '';
  try {
    await login(email.value, password.value);
    router.push('/user/monthly-count-report');
  } catch (error) {
    loginError.value = error instanceof Error ? error.message : String(error);
  }
};
</script>
