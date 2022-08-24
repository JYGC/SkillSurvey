<template>
  <div class="row">
    <div class="col-sm-2 nav-div">
      <b-nav vertical>
        <b-nav-item to="/" active>Home</b-nav-item>
        <b-nav-item to="/monthly-count-report">Monthly count report</b-nav-item>
        <b-nav-item class="expand">
          <b-button v-b-toggle.collapse-skill-details variant="primary" class="w-100">Skill details</b-button>
          <b-collapse id="collapse-skill-details">
            <b-nav-item to="/skill-list">Skills list</b-nav-item>
            <b-nav-item to="/skill-type-list">Skill type list</b-nav-item>
          </b-collapse>
        </b-nav-item>
      </b-nav>
    </div>
    <div class="col-sm-10">
      <router-view />
    </div>
  </div>
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
}

.nav-div {
  padding-top: 60px;
  padding-right: 0px;
}

.nav-div, .nav-div .dropdown-menu {
  background-color: darkslateblue;
}

.nav-div .nav-item:not(.expand) a:hover {
  background-color: lightslategray;
}

.nav-div .nav-item:not(.expand) a.router-link-exact-active {
  background-color: slateblue;
}

.expand button:not(.collapsed) {
  font-weight: bold;
}

.expand .nav-link {
  padding: 0;
}

.expand .nav-link:hover {
  background-color: rgb(83, 93, 104);
}

.nav-div .nav-item.expand button {
  background-color: transparent;
  border: hidden;
  border-radius: 0;
}

.nav-div .nav-item.expand button:focus {
  outline: none;
  box-shadow: none;
}

.nav-div .nav-item:not(.expand) a, .nav-div .nav-item button {
  color: white;
  padding: 20px 0px 20px 0px;
}

#app > .row {
  height: 98vh;
}
</style>
