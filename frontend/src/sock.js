import store from './store';

const reConnectTimeout = 2000;

// Simple re-connecting websocket

let ws = null;

function connect(host) {
    const url = `ws://${host}/ws`;

    ws = new WebSocket(url);

    ws.onopen = function () {
        console.log('opened ws');
        store.commit('updateConnectionStatus', true);
        store.commit('setWebsocketClient', ws);
    };

    ws.onerror = function (e) {
        console.error('got ws error', e);
        store.commit('updateConnectionStatus', false);
        ws.close();
    };

    ws.onclose = function (e) {
        console.warn('closed ws', e);
        store.commit('updateConnectionStatus', false);

        setTimeout(function() {
            connect(host);
        }, reConnectTimeout)
    };

    ws.onmessage = function (msg) {
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
                // store.commit('addNewChatMessage', obj);
                store.commit('acknowledgeSubscription', obj.Body);
                break;
            case 'CHAT':
                console.log("Got CHAT", obj);
                store.commit('addNewChatMessage', obj.Body);
                break;
            case 'CHATACK':
                console.log("Got CHATACK", obj);
                const roomId = obj.Body.Channel.split('.').pop();
                store.commit('acknowledgeChatMessage', {roomId, messageClientId: obj.Body.ClientID});
                break;
            default:
                console.log('Got other msg', obj);
                break;
        }
    };
}

export {
    ws,
    connect,
};
