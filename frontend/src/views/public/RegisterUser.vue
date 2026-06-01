<template>
  <div>
    <CvFluidForm>
      <CvTextInput label="Name" v-model="name" />
      <CvTextInput label="Email" v-model="email" />
      <CvTextInput label="Password" type="password" v-model="password" />
      <CvTextInput label="Confirm Password" type="password" v-model="confirmPassword" />
    </CvFluidForm>
    <br />
    <CvButton @click="onSubmit()">Register</CvButton>
    <p v-if="registerError" data-testid="register-error">{{ registerError }}</p>
    <br />
    <br />
    <p>Already have an account?</p>
    <CvLink href="/login">Login</CvLink>
  </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { authRepository } from '@/repositories/auth.repository';

const router = useRouter();

const name = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');
const registerError = ref('');

const onSubmit = async () => {
  registerError.value = '';
  if (password.value !== confirmPassword.value) {
    registerError.value = 'Passwords do not match!';
    return;
  }
  try {
    await authRepository.register(name.value, email.value, password.value, confirmPassword.value);
    router.push('/');
  } catch (error) {
    registerError.value = error instanceof Error ? error.message : String(error);
  }
};
</script>
