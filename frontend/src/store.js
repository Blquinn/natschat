import Vue from 'vue';
import Vuex from 'vuex';
import HttpClient from "./httpclient";
import {connect} from "./sock";

Vue.use(Vuex);

export function mapChatRoom(room) {
  return {
    id: room.id,
    insertedAt: room.insertedAt,
    updatedAt: room.updatedAt,
    name: room.name,
    active: false,
    loading: false,
    channelSubscribed: false,
    chatLog: [],
    channel: `chat.room.${room.id}`,
  };
}

export function mapChatMessage(message) {
  return {
    id: message.id,
    clientId: '',
    content: message.body,
    user: message.user,
    acknowledged: true,
    deliveryFailure: false,
  };
}

function createUID() {
  return new Date().getTime().toString() + Math.random().toString().substr(2, 9);
}

export default new Vuex.Store({
  state: {
    ws: null,
    httpClient: null,
    user: null,
    socketConnected: false,
    socket: null,
    chatRooms: [],
    loadingChatRooms: false,
  },

  getters: {
    openRooms(state) {
      return state.chatRooms.filter(r => r.active === true);
    }
  },

  mutations: {
    updateConnectionStatus(state, connected) {
      state.socketConnected = connected;
    },

    setWebsocketClient(state, ws) {
      state.ws = ws;
    },

    setHttpClient(state, client) {
      state.httpClient = client;
    },

    setUser(state, user) {
      state.user = user;
    },

    addNewChatMessage(state, {roomId, message}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        const msg = room.chatLog.find(m => m.id === message.id);
        if (msg === undefined) {
          room.chatLog.push(message);
        }
      }
    },

    acknowledgeMessageDelivery(state, {roomId, messageId}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        const message = room.chatLog.find(m => m.id === messageId);
        if (message !== undefined) {
          message.acknowledged = true;
        }
      }
    },

    acknowledgeSubscription(state, subscription) {
      const roomID = subscription.channel.split('.').pop();

      const room = state.chatRooms.find(r => r.id === roomID);
      if (room === undefined) {
        console.error(`Chat room with id ${roomID} not found, while acknowledging subscription.`);
        return;
      }

      room.channelSubscribed = true;
    },

    acknowledgeChatMessage(state, {roomId, messageClientId, messageId}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room === undefined) {
        console.error(`Chat room with id ${roomId} not found, while acknowledging chat message.`);
        return;
      }

      const msg = room.chatLog.find(m => m.clientId === messageClientId);
      if (msg === undefined) {
        console.error(`Chat message with clientId ${messageClientId} not found in room ${roomId}, while acknowledging chat message.`);
        return;
      }

      msg.id = messageId;
      msg.acknowledged = true;
    },

    addChatRoom(state, room) {
      state.chatRooms.push(room);
    },

    syncChatRooms(state, rooms) {
      state.chatRooms = rooms;
    },

    setLoadingRooms(state, isLoading) {
      state.loadingChatRooms = isLoading;
    },

    setLoadingRoom(state, {roomId, isLoading}) {
      console.info(`Setting room ${roomId} loading state to ${isLoading}`);
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        room.loading = isLoading;
      }
    },

    openChatRoom(state, roomId) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        room.active = true;
      }
    },

    setChatLog(state, {roomId, log}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        room.chatLog = log;
      }
    },

    checkMessageAcknowledged(state, {roomId, messageClientId}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room === undefined) {
        console.error(`Room with id ${roomId} not found while acknowledging message.`);
        return;
      }

      const msg = room.chatLog.find(l => l.clientId === messageClientId);
      if (msg === undefined) {
        console.error(`Message with clientId ${messageClientId} not found in room ${roomId} while acknowledging message.`);
        return;
      }

      if (msg.acknowledged !== true) {
        console.error(`Message with clientId ${messageClientId} in room ${roomId} was not acknowledged and is being marked failed.`);
        msg.deliveryFailure = true;
      }
    },

  },


  actions: {
    loginAndConnect: function(context, {host, username, password}) {
      const client = new HttpClient(host, '');
      client.post('/login', {username, password})
        .catch(err => {
          alert('Login failed');
          console.error('Login failed', err);
        }).then(res => {
          context.commit('setHttpClient', new HttpClient(host, res.data.token));
          context.commit('setUser', res.data.user);
          connect(host, res.data.token);
        });
    },
    loadChatRooms: function (context) {
      context.commit('setLoadingRooms', true);
      context.state.httpClient.get('/api/rooms')
        .catch(err => {
          alert('An error occurred while loading chat rooms');
          console.error(err);
        })
        .then(res => {
          this.commit('syncChatRooms', res.data.results.map(r => mapChatRoom(r)));
          context.commit('setLoadingRooms', false);
        });
    },
    createChatRoom: function(context, name) {
      if (name === '') {
        return;
      }

      let body = {name: name};

      context.state.httpClient.post('/api/rooms', body)
        .catch(err => {
          alert('An error occurred while creating chat room');
          console.error(err);
        }).then(res => {
          context.commit('addChatRoom', mapChatRoom(res.data));
        });
    },
    getChatHistoryThenSubscribe: function(context, {roomId, channel}) {
      context.state.httpClient.get(`/api/rooms/${roomId}/history`)
        .catch(err => {
          alert('Error while retrieving chat history');
          console.error(err)
        }).then(res => {
          console.info(res.data);
          const chatLog = res.data.results.map(msg => mapChatMessage(msg));
          context.commit('setChatLog', {
            roomId: roomId,
            log: chatLog,
          });
          context.commit('setLoadingRoom', {roomId: roomId, isLoading: false});
          context.dispatch('subscribeToChannel', channel);
        })
    },
    subscribeToChannel: function(context, channel) {
      context.state.ws.send(JSON.stringify({
        type: 'SUB',
        body: {
          channel: channel,
        }
      }));
    },
    sendMessage: function(context, {room, messageBody}) {
      if (context.state.ws === null) {
        alert('Websocket not connected');
        return;
      }

      let clientId = createUID();

      context.commit('addNewChatMessage', {
        roomId: room.id,
        message: {
          content: messageBody,
          user: context.state.user,
          acknowledged: false,
          clientId: clientId,
          deliveryFailure: false,
        },
      });

      context.state.ws.send(JSON.stringify({
        type: 'CHAT',
        body: {
          channel: room.channel,
          content: messageBody,
          clientId: clientId,
        }
      }));

      // If the message is not acknowledged withing 5s, mark it not delivered
      setTimeout(function () {
        context.commit('checkMessageAcknowledged', {
          roomId: room.id,
          messageClientId: clientId,
        });
      }, 5000)
    }
  }
})
