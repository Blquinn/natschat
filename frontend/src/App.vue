<template>
    <div id="app">
        <div class="connection-indicator">
            <span v-if="!connected" style="background-color: orange">Socket not connected</span>
            <span v-else style="background-color: green">Socket connected</span>
        </div>

        <div id="main" v-if="connected">
            <router-view/>
        </div>
        <!--TODO: Move form into component-->
        <div id="setup" v-else v-on:keyup.enter="startConnection()">
            <div>
                <label for="host-txt">Host: </label>
                <input id="host-txt" type="text" v-model="host" />
            </div>

            <div>
                <label for="username-txt">Username: </label>
                <input id="username-txt" type="text" v-model="username" />
            </div>

            <div>
                <label for="password-txt">Password: </label>
                <input id="password-txt" type="text" v-model="password" />
            </div>
            <button v-on:click="startConnection()">Connect</button>
        </div>
    </div>
</template>

<script>
  export default {
    computed: {
      connected() {
        return this.$store.state.socketConnected;
      },
    },
    data: function() {
      return {
        host: 'localhost:5000',
        username: 'ben',
        password: 'password',
      }
    },
    methods: {
      startConnection() {
        if (this.host === '') {
          console.log('No host entered');
          return;
        }

        console.log(`Setting host to ${this.host}`);
        const host = this.host;
        const username = this.username;
        const password = this.password;
        this.$store.dispatch('loginAndConnect', {host, username, password});
      }
    }
  };
</script>

<style scoped>
</style>
