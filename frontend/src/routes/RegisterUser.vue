<template>
  <CvFluidForm>
    <CvTextInput label="Name" v-model="name" />
    <CvTextInput label="Email" v-model="email" />
    <CvPasswordInput label="Password" v-model="password" />
    <CvPasswordInput label="Confirm Password" v-model="confirmPassword" />
  </CvFluidForm>
  <br />
  <CvButton>Register</CvButton>
</template>
<script lang="ts" setup>
import { ref } from 'vue';

const name = ref('');
const email = ref('');
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
</script>