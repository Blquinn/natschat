<template>
    <div id="app">
        <div class="connection-indicator">
            <span v-if="!connected" style="background-color: orange">Socket not connected</span>
            <span v-else style="background-color: green">Socket connected</span>
        </div>

        <div id="main" v-if="connected">
            <router-view/>
        </div>
        <div id="setup" v-else>
            <input type="text" v-model="host" v-on:keyup.enter="startConnection()"/>
            <button v-on:click="startConnection()">Connect</button>
        </div>
    </div>
</template>

<script>
  import HttpClient from "./httpclient";
  import {connect} from './sock';

  export default {
    computed: {
      connected() {
        return this.$store.state.socketConnected;
      },
    },
    data: function() {
      return {
        host: 'localhost:5000',
      }
    },
    methods: {
      startConnection() {
        if (this.host === '') {
          console.log('No host entered');
          return;
        }

        console.log(`Setting host to ${this.host}`);
        connect(this.host);
        this.$store.commit('setHttpClient', new HttpClient(this.host));
      }
    }
  };
</script>

<style scoped>
</style>
