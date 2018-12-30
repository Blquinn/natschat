import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex);

const chatRoom = [
  {
    ID: '',
    chatLog: [
      {
        ID: '',
        Body: '',
        User: {
          Username: '',
        }
      }
    ]
  }
];

export default new Vuex.Store({
  // strict: process.env.NODE_ENV !== 'production',
  state: {
    socketConnected: false,
    socket: null,
    chatRooms: [],
  },
  mutations: {
    updateConnectionStatus(state, connected) {
      state.socketConnected = connected;
    },

    addNewChatMessage(state, message) {
      const room = state.chatRooms.find(r => r.ID === message.RoomID);
      if (room !== undefined) {
        room.chatLog.push(message);
      }
    },

    acknowledgeSubscription(state, subscription) {
      const roomID = subscription.Channel.split('.').pop();

      const room = state.chatRooms.find(r => r.ID === roomID);
      if (room === undefined) {
        console.error(`Chat room with id ${roomID} not found, while acknowledging subscription.`);
        return;
      }

      // const msg = room.chatLog.find(m => m.ID === message.ID);
      // msg.acknowledged = true;

    },

    acknowledgeChatMessage(state, message) {
      const roomID = message.Channel.split('.').pop();

      const room = state.chatRooms.find(r => r.ID === roomID);
      if (room === undefined) {
        console.error(`Chat room with id ${roomID} not found, while acknowledging chat message.`);
        return;
      }

      const msg = room.chatLog.find(m => m.ID === message.ID);
      msg.acknowledged = true;
    },

    addChatRoom(state, room) {
      state.chatRooms.push(room);
    },
  },
  actions: {

  }
})
