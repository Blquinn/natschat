<template>
    <div class="chat-container">
        <div class="active" v-if="!room.loading">
            <h1>{{room.name}}</h1>
            <div id="chat-log" v-chat-scroll="{always: false, smooth: true}">
                <div v-for="message in room.chatLog" :key="message.id" class="message-container"
                     v-bind:class="{'right': message.user.id === user.id, 'left': message.user.id !== user.id}">
                    <div>
                        <span>{{ message.user.username }}</span>:
                        <span>{{ message.content }}</span>
                        <span v-if="message.acknowledged">✔️</span>
                        <span v-if="message.deliveryFailure">❌ - Failed to send message</span>
                    </div>
                </div>
            </div>

            <div id="chat-input">
                <input type="text" v-model="messageInput" v-on:keyup.13="sendMessage">
                <button id="send-message-btn" v-on:click="sendMessage">Send</button>
            </div>
        </div>
        <circle2 v-else></circle2>
    </div>
</template>

<script>
  import Circle2 from 'vue-loading-spinner/src/components/Circle2';

  export default {
    name: "ChatContainer",
    components: {
      Circle2,
    },
    props: {
      room: Object,
    },
    data: function () {
      return {
        user: this.$store.state.user,
        messageInput: '',
      };
    },
    created: function () {
      this.$store.dispatch('getChatHistoryThenSubscribe', {roomId: this.room.id, channel: this.room.channel});
    },
    methods: {
      sendMessage: function sendMessage() {
        if (this.messageInput === '') {
          return;
        }

        this.$store.dispatch('sendMessage', {
          room: this.room,
          messageBody: this.messageInput,
        });

        this.messageInput = '';
      },
    }
  }
</script>

<style scoped>
    .chat-container {
        display: inline-block;
        margin-right: 2em;
    }

    #chat-log {
        height: 20em;
        width: 30em;
        overflow-y: scroll;
    }

    .message-container + .right {
        text-align: right;
    }
</style>
