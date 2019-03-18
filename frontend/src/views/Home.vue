<template>
  <div class="home">
    <!--<input type="text" >-->
    <div class="chat-rooms-container">
      <p>Select a chat room</p>
      <div class="room" v-for="room in rooms" :key="room.id" v-on:click="openChatRoom(room)">
        {{ room.name }}
      </div>
    </div>

    <div class="create-chat-room-container">
      <p>Or, create one</p>
      <input type="text" placeholder="Room name"
             v-model="roomNameInput"
             v-on:keyup.enter="createChatRoom()" />
      <button v-on:click="createChatRoom()">Create</button>
    </div>

    <div class="chats-container">
        <ChatContainer v-for="room in openRooms" v-bind:room="room" :key="room.id" />
    </div>

  </div>
</template>

<script>
// @ is an alias to /src
import ChatContainer from "../components/ChatContainer";

export default {
  name: 'home',
  computed: {
    rooms() {
      return this.$store.state.chatRooms
    },
    openRooms() {
      return this.$store.getters.openRooms;
    },
  },
  data: function() {
    return {
      roomNameInput: '',
    };
  },
  created: function() {
    this.$store.dispatch('loadChatRooms')
  },
  methods: {
    createChatRoom: function() {
      if (this.roomNameInput === '') {
        return;
      }

      this.$store.dispatch('createChatRoom', this.roomNameInput);

      this.roomNameInput = '';
    },
    openChatRoom: function(room) {
      this.$store.commit('openChatRoom', room.id);
    },
  },
  components: {
    ChatContainer
  }
}
</script>

<style scoped>
  .room {
    cursor: pointer;
    display: inline-block;
    margin: .5em;
    padding: .5em;
    background-color: aquamarine;
  }

  .chats-container {
    display: inline-block;
  }
</style>
