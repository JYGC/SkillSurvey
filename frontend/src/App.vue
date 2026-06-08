<template>
  <router-view />
</template>

<script lang="ts" setup>
</script>

<style lang="scss">
  @import './styles/carbon';
</style>
<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.fill-parent {
  width: 100%;
  height: 100%;
}

.vertical-padding {
    margin: 10px 0 10px 0;
}

tbody {
    display: block;
    height: calc( 100vh - 100px );
    overflow: auto;
}
thead, tbody tr {
    display: table;
    width: 100%;
    table-layout: fixed;
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
</style>
