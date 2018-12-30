<template>
  <div class="home">
    <!--<input type="text" >-->
    <div class="chat-rooms-container">
      <p>Select a chat room</p>
      <div class="room" v-for="room in rooms" v-on:click="openChatRoom(room)">
        {{ room.Name }}
      </div>
    </div>

    <div class="create-chat-room-container">
      <p>Or, create one</p>
      <input type="text" placeholder="Room name"
             v-model="roomNameInput"
             v-on:keyup.enter="createChatRoom()" />
      <button v-on:click="createChatRoom()">Create</button>
    </div>

    <div class="chat-container">
      <ChatContainer v-for="room in openRooms" v-bind:room-id="room.ID" />
    </div>
  </div>
</template>

<script>
// @ is an alias to /src
import ChatContainer from "../components/ChatContainer";

import http from '../httpclient';

export default {
  name: 'home',
  data: function() {
    return {
      loadingRooms: true,
      rooms: [],
      openRooms: [],
      roomNameInput: '',
    };
  },
  created: function() {
    this.loadChatRooms()
  },
  methods: {
    loadChatRooms: function() {
      http.get('/api/rooms')
      .catch(err => {
        alert('An error occurred while loading chat rooms');
        console.error(err);
      }).then(res => {
        this.rooms = res.data.Results;
        this.loadingRooms = false;
      })
    },
    createChatRoom: function() {
      if (this.roomNameInput === '') {
        return;
      }

      let body = {Name: this.roomNameInput};

      http.post('/api/rooms', body)
      .catch(err => {
        alert('An error occurred while creating chat room');
        console.error(err);
      }).then(res => {
        this.rooms.push(res.data);
      });

      this.roomNameInput = '';
    },
    openChatRoom: function(room) {
      this.openRooms.push(room)
    },
  },
  components: {
    ChatContainer
  }
}
</script>

<style scoped>
  .room {
    margin-top: .15em;
    cursor: pointer;
  }
</style>
