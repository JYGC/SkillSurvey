<template>
  <CvFluidForm>
    <CvTextInput label="Name" v-model="name" />
    <CvTextInput label="Email" v-model="email" />
    <CvTextInput label="Password" type="password" v-model="password" />
    <CvTextInput label="Confirm Password" type="password" v-model="confirmPassword" />
  </CvFluidForm>
  <br />
  <CvButton @click="onSubmit()">Register</CvButton>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import PocketBase from 'pocketbase';

const pb = new PocketBase('http://192.168.8.145:8090');

const name = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');

function validateForm() {
  if (password.value !== confirmPassword.value) {
    alert('Passwords do not match!');
    return false;
  }
  // Add more validation as needed
  return true;
}

async function registerUser() {
  try {
    const record = await pb.collection('users').create({
      name: name.value,
      email: email.value,
      password: password.value,
      passwordConfirm: confirmPassword.value,
    });

    console.log('User registered:', record);
  } catch (error) {
    console.error('Error registering user:', error);
    alert('Failed to register user. Please try again.');
  }
}

function onSubmit() {
  if (!validateForm()) {
    return;
  }
  registerUser();
}
</script>