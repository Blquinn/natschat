<template>
  <div id="app">
    <!--<div id="nav">-->
      <!--<router-link to="/">Home</router-link>-->
    <!--</div>-->
    <router-view/>
  </div>
</template>

<script>
  export default {
    data: function() {
      return {
        ws: null,
        connected: false,
      }
    },
    created: function () {
      // const vue = this;
      // var ws = new WebSocket('ws://localhost:5000/ws');
      // ws.onopen = function () {
      //   console.log('opened ws');
      //   this.connected = true;
      // };
      //
      // ws.onerror = function (e) {
      //   console.error('got ws error');
      //   this.connected = false;
      // };
      //
      // ws.onclose = function (e) {
      //   console.warn('closed ws', e);
      //   this.connected = false;
      // };
      //
      // ws.onmessage = function (msg) {
      //   // console.info('got ws msg', msg);
      //   let obj;
      //   try {
      //     obj = JSON.parse(msg.data);
      //   } catch (error) {
      //     console.error(error, msg);
      //     return;
      //   }
      //
      //   switch (obj.Type) {
      //     case 'SUBACK':
      //       console.log('Got SUBACK', obj.Body);
      //       break;
      //     case 'CHAT':
      //       console.log("Got CHAT", obj);
      //       break;
      //     case 'CHATACK':
      //       console.log("Got CHATACK", obj);
      //       const logMsg = vue.chatLog.find(l => l.clientId === obj.Body.ClientID);
      //       if (logMsg !== undefined) {
      //         logMsg.acknowledged = true;
      //       }
      //       break;
      //     default:
      //       console.log('Got other msg', obj);
      //       break;
      //   }
      // };
      // this.ws = ws;
    }
  };
</script>

<style scoped>
</style>
