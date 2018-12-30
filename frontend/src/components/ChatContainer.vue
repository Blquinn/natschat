<template>
    <div id="app">
        <h1>Chat</h1>
        <div class="connection-indicator">
            <span v-if="connected" style="background-color: orange">Not connected</span>
            <span v-else style="background-color: green">Connected</span>
        </div>
        <div id="chat-log">
            <div v-for="message in chatLog">
                <span>{{ message.user }}</span>:
                <span>{{ message.content }}</span>
                <span v-if="message.acknowledged">✔️</span>
                <span v-if="message.deliveryFailure">❌ - Failed to send message</span>
            </div>
        </div>

        <div id="chat-input">
            <input type="text" v-model="messageInput" v-on:keyup.13="sendMessage">
            <button id="send-message-btn" v-on:click="sendMessage">Send</button>
        </div>
    </div>
</template>

<script>
import axios from 'axios';

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

export default {
    name: "ChatContainer",
    props: {
        roomId: String,
    },
    data: function() {
        return {
            channel: `chat.rooms.${this.roomId}`,
            chatLog: [],
            ws: null, // WebSocket
            connected: false,
            messageInput: '',
        };
    },
    created: function() {
        const vue = this;
        var ws = new WebSocket('ws://localhost:5000/ws');
        ws.onopen = function () {
            console.log('opened ws');
            this.connected = true;
            vue.subscribeToChannel(vue.channel);
        };

        ws.onerror = function (e) {
            console.error('got ws error');
            this.connected = false;
        };

        ws.onclose = function (e) {
            console.warn('closed ws', e);
            this.connected = false;
        };

        ws.onmessage = function (msg) {
            // console.info('got ws msg', msg);
            let obj;
            try {
                obj = JSON.parse(msg.data);
            } catch (error) {
                console.error(error, msg);
                return;
            }

            switch (obj.Type) {
                case 'SUBACK':
                    console.log('Got SUBACK', obj.Body);
                    break;
                case 'CHAT':
                    console.log("Got CHAT", obj);
                    break;
                case 'CHATACK':
                    console.log("Got CHATACK", obj);
                    const logMsg = vue.chatLog.find(l => l.clientId === obj.Body.ClientID);
                    if (logMsg !== undefined) {
                        logMsg.acknowledged = true;
                    }
                    break;
                default:
                    console.log('Got other msg', obj);
                    break;
            }
        };
        this.ws = ws;

        this.getChatHistory();
    },
    methods: {
        getChatHistory: function() {
            axios.get(`http://localhost:5000/api/rooms/${this.roomId}/history`, {
                headers: {
                    Authorization: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJxdWlubkBtYXRhZG9yYXBwLmNvbSIsImV4cCI6MjU0NjEzNTk4NCwidXNlcl9pZCI6ImU0OTE4OTgzLWY4YzEtNGE0YS1iODE4LWQ0YjMxMTQ5ZDZjNCIsInVzZXJuYW1lIjoiYmVuIn0.ZxMyCa03yitGrpLK3ZUZv490YAzERrVVnkVq-SoMIDU`
                }
            }).catch(err => {
                alert('Error while retrieving chat history');
                console.error(err)
            }).then(res => {
                console.info(res.data);
                this.chatLog = res.data.Results.map(msg => mapChatMessage(msg))
            })
        },
        sendMessage: function sendMessage() {
            if (this.messageInput === '' && this.ws !== null) {
                return;
            }

            if (this.ws === null) {
                alert('Websocket not connected');
                return;
            }

            let clientId = this.createUID();

            this.chatLog.push({
                content: this.messageInput,
                user: 'ben',
                acknowledged: false,
                clientId: clientId,
                deliveryFailure: false,
            });

            this.ws.send(JSON.stringify({
                Type: 'CHAT',
                Body: {
                    Channel: this.channel,
                    Content: this.messageInput,
                    ClientID: clientId,
                }
            }));

            this.messageInput = '';

            const vue = this;
            // If the message is not acknowledged withing 10s, mark it not delivered
            setTimeout(function () {
                const logMsg = vue.chatLog.find(l => l.clientId === clientId);
                if (logMsg !== undefined && logMsg.acknowledged !== true) {
                    logMsg.deliveryFailure = true;
                }
            }, 10000)
        },
        subscribeToChannel: function subscribeToChannel(channel) {
            this.ws.send(JSON.stringify({
                Type: 'SUB',
                Body: {
                    Channel: channel,
                }
            }));
        },
        createUID: function createUID() {
            return new Date().getTime().toString() + Math.random().toString().substr(2, 9)
        }
    }
}
</script>

<style scoped>
    #chat-log {
        height: 20em;
        width: 34em;
        overflow-y: scroll;
    }
</style>
