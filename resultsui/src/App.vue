<template>
  <!-- <div id="nav">
      <router-link to="/">Home</router-link> |
      <router-link to="/monthly-count-report">Monthly count report</router-link>
  </div>
  <router-view /> -->
  <MonthlyCountReport />
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import MonthlyCountReport from './components/MonthlyCountReport.vue';

export default defineComponent({
  name: 'App',
  components: {
    MonthlyCountReport
  }
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
