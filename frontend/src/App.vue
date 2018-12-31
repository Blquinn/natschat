<template>
  <div id="app">
    <div class="connection-indicator">
        <span v-if="!connected" style="background-color: orange">Socket not connected</span>
        <span v-else style="background-color: green">Socket connected</span>
    </div>
    <!--</div>-->
    <!--<div id="nav">-->
      <!--<router-link to="/">Home</router-link>-->
    <!--</div>-->
    <router-view/>
  </div>
</template>

<script>
  export default {
    computed: {
      connected() {
          return this.$store.state.socketConnected
      }
    },
  };
</script>

<style scoped>
</style>
