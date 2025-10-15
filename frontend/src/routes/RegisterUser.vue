<template>
  <CvFluidForm>
    <CvTextInput label="Name" v-model="name" />
    <CvTextInput label="Email" v-model="email" />
    <CvTextInput label="Password" type="password" v-model="password" />
    <CvTextInput label="Confirm Password" type="password" v-model="confirmPassword" />
  </CvFluidForm>
  <br />
  <CvButton>Register</CvButton>
</template>
<script lang="ts" setup>
import { ref } from 'vue';

const name = ref('');
const email = ref('');
const password = ref('');
const confirmPassword = ref('');
</script>