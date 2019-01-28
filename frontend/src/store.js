import Vue from 'vue';
import Vuex from 'vuex';
import HttpClient from "./httpclient";
import {connect} from "./sock";

Vue.use(Vuex);

function mapChatRoom(room) {
  return {
    id: room.ID,
    insertedAt: room.InsertedAt,
    updatedAt: room.UpdatedAt,
    name: room.Name,
    active: false,
    loading: false,
    channelSubscribed: false,
    chatLog: [],
    channel: `chat.room.${room.ID}`,
  };
}

function mapChatMessage(message) {
  return {
    id: message.ID,
    clientId: '',
    content: message.Body,
    user: message.User.Username,
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
    socketConnected: false,
    socket: null,
    chatRooms: [],
    loadingChatRooms: false,
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

    addNewChatMessage(state, {roomId, message}) {
      const room = state.chatRooms.find(r => r.id === roomId);
      if (room !== undefined) {
        room.chatLog.push(message);
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
      const roomID = subscription.Channel.split('.').pop();

      const room = state.chatRooms.find(r => r.id === roomID);
      if (room === undefined) {
        console.error(`Chat room with id ${roomID} not found, while acknowledging subscription.`);
        return;
      }

      room.channelSubscribed = true;
    },

    acknowledgeChatMessage(state, {roomId, messageClientId}) {
      // const roomID = message.Channel.split('.').pop();

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
          connect(host);
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
          this.commit('syncChatRooms', res.data.Results.map(r => mapChatRoom(r)));
          context.commit('setLoadingRooms', false);
        });
    },
    createChatRoom: function(context, name) {
      if (name === '') {
        return;
      }

      let body = {Name: name};

      context.state.httpClient.post('/api/rooms', body)
        .catch(err => {
          alert('An error occurred while creating chat room');
          console.error(err);
        }).then(res => {
          context.commit('addChatRoom', mapChatRoom(res.data));
        });
    },
    getChatHistoryThenSubscribe: function(context, room) {
      context.state.httpClient.get(`/api/rooms/${room.id}/history`)
        .catch(err => {
          alert('Error while retrieving chat history');
          console.error(err)
        }).then(res => {
          console.info(res.data);
          const chatLog = res.data.Results.map(msg => mapChatMessage(msg));
          context.commit('setChatLog', {
            roomId: room.id,
            log: chatLog,
          });
          context.commit('setLoadingRoom', {roomId: room.id, isLoading: false});
          context.dispatch('subscribeToChannel', room.channel);
        })
    },
    subscribeToChannel: function(context, channel) {
      context.state.ws.send(JSON.stringify({
        Type: 'SUB',
        Body: {
          Channel: channel,
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
          user: 'ben',
          acknowledged: false,
          clientId: clientId,
          deliveryFailure: false,
        },
      });

      context.state.ws.send(JSON.stringify({
        Type: 'CHAT',
        Body: {
          Channel: room.channel,
          Content: messageBody,
          ClientID: clientId,
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
