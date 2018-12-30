<template>
    <div id="chat-container">
        <div class="active" v-if="active">
            <h1>Chat</h1>
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
        <!--<circle v-else></circle>-->
        <circle2 v-else></circle2>
    </div>
</template>

<script>
import ws from '../sock';
import http from '../httpclient';
import Circle2 from 'vue-loading-spinner/src/components/Circle2';

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
    components: {
        Circle2,
    },
    props: {
        roomId: String,
    },
    data: function() {
        return {
            channel: `chat.rooms.${this.roomId}`,
            chatLog: [],
            messageInput: '',
            active: false,
        };
    },
    created: function() {
        // this.subscribeToChannel(this.channel);
        this.getChatHistoryThenSubscribe();
        // this.subscribeToChannel(this.channel);
    },
    methods: {
        getChatHistoryThenSubscribe: function() {
            http.get(`/api/rooms/${this.roomId}/history`)
            .catch(err => {
                alert('Error while retrieving chat history');
                console.error(err)
            }).then(res => {
                console.info(res.data);
                this.chatLog = res.data.Results.map(msg => mapChatMessage(msg));
                this.subscribeToChannel(this.channel);
            })
        },
        sendMessage: function sendMessage() {
            if (this.messageInput === '' && ws !== null) {
                return;
            }

            if (ws === null) {
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

            ws.send(JSON.stringify({
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
            ws.send(JSON.stringify({
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
