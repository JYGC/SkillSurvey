<template>
  <div class="row">
    <div class="col-md-2 nav-div">
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
    <div class="col-md-10">
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

.nav > .association * {
  background-color: rgb(131, 180, 221);
  color: white;
}

.nav > .association:nth-child(even) * {
  background-color: steelblue;
  color: white;
}

.nav > .new-association * {
  background-color: rgb(70, 176, 180);
  color: white;
}

.nav > .association:hover *, .nav > .new-association:hover * {
  background-color: rgb(180, 207, 230);
  color: #0d6efd;
}

.margin-left-10 {
  margin-left: 10px;
}

.nav-div {
  padding-top: 60px;
  padding-right: 0px;
}

.nav-div, .nav-div .dropdown-menu {
  background-color: darkslateblue;
}

.nav-div .nav-item:not(.expand) a:hover, .btn:hover {
  background-color: lightslategray;
}

.nav-div .nav-item:not(.expand) a.router-link-exact-active {
  background-color: slateblue;
}

.fill-parent {
  width: 100%;
  height: 100%;
}

.expand button:not(.collapsed) {
  font-weight: bold;
}

.btn {
    border: hidden;
    border-radius: 0;
    text-decoration-line: none;
    background-color: darkslateblue;
}

.btn-danger {
  background-color: red;
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

.nav-div .nav-item.expand button:focus, .btn:focus {
  outline: none;
  box-shadow: none;
}

.nav-div .nav-item:not(.expand) a, .nav-div .nav-item button {
  color: white;
  padding: 20px 0px 20px 0px;
}

#app > .row {
  height: 100vh;
  width: 100vw;
}

/* Affects SkillTypeList needs to be in common file */
tbody {
    display: block;
    height: calc( 100vh - 100px );
    overflow: auto;
}
thead, tbody tr {
    display: table;
    width: 100%;
    table-layout: fixed;/* even columns width , fix width of table too*/
}
table {
    width: 100%;
}

thead tr {
    font-weight: bold;
    background-color: darkslateblue;
    color: white;
}

table tbody tr:nth-child(even) {
    background-color: rgb(239, 240, 248);
}

.vertical-padding {
    margin: 10px 0 10px 0;
}
</style>
