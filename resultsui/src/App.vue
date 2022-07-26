<template>
  <div id="nav">
      <router-link to="/">Home</router-link> |
      <router-link to="/monthly-count-report">Monthly count report</router-link> |
      <router-link to="/skill-list">Skills list</router-link> |
      <router-link to="/skill-type-list">Skill type list</router-link>
  </div>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'App'
});
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
