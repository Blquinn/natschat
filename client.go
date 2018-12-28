package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/davecgh/go-spew/spew"

	"gopkg.in/go-playground/validator.v9"

	"github.com/gorilla/websocket"
)

func init() {
	gob.Register(ChatMessage{})
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// SendJson pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	gnats *Gnats

	// subs map[string]*nats.Subscription
	subs map[*Subscription]bool

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func newClient(hub *Hub, gnats *Gnats, conn *websocket.Conn) *Client {
	return &Client{
		hub:   hub,
		gnats: gnats,
		subs:  make(map[*Subscription]bool),
		conn:  conn,
		send:  make(chan []byte, 100),
	}
}

func (c *Client) handleMessage(bts []byte) {
	var raw json.RawMessage
	m := Message{
		Body: &raw,
	}
	err := json.Unmarshal(bts, &m)
	if err != nil {
		c.send <- []byte("invalid msg")
		return
	}

	i, f := TypeMap[m.Type]
	if !f {
		c.send <- []byte("invalid msg type")
		return
	}

	msg := i()
	err = json.Unmarshal(raw, &msg)
	if err != nil {
		c.send <- []byte("invalid msg body")
	}

	err = validate.Struct(msg)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			log.Debugf("error while validating struct: %v", err)
			c.send <- []byte("internal server error")
			return
		}

		ves := make([]ValidationError, len(errs))
		for i, e := range errs {
			ves[i] = ValidationError{
				Field:   e.Field(),
				Message: fmt.Sprintf("%v", e),
			}
			e.Field()
		}

		spew.Dump(errs)

		m.Body = msg
		errMsg := Message{
			Type: MessageTypeValidationErr,
			Body: ValidationErrorMessage{
				OriginalMessage: string(bts),
				Errors:          ves,
			},
		}

		j, err := json.Marshal(errMsg)
		if err != nil {
			log.Errorln("error while marshalling validation errs")
			c.send <- []byte("internal server error")
			return
		}

		c.send <- j
		return
	}

	switch msg := msg.(type) {
	case *SubscriptionMessage:
		if m.Type == MessageTypeSub {
			c.handleSubMessage(msg)
		} else {
			c.handleUnSubMessage(msg)
		}
	case *ChatMessage:
		c.handleChatMessage(msg)
	default:
		c.send <- []byte(fmt.Sprintf("got unhandled type: %T", msg))
	}
}

func (c *Client) SendJSON(msg interface{}) {
	b, err := json.Marshal(msg)
	if err != nil {
		errMsg, _ := json.Marshal(Message{
			Type: MessageTypeServerErr,
			Body: "failed to serialize response",
		})
		c.send <- []byte(errMsg)
		log.Errorln("failed to serialize response: " + err.Error())
		return
	}
	c.send <- b
}

func (c *Client) handleSubMessage(m *SubscriptionMessage) {
	for s := range c.subs {
		if s.sub.Subject == m.Channel {
			// User is already subscribed to the channel
			r := Message{
				Type: MessageTypeSubAck,
				Body: m,
			}
			c.SendJSON(r)
			return
		}
	}

	sub, err := c.gnats.Subscribe(m.Channel, c)
	if err != nil {
		r := Message{
			Type: MessageTypeServerErr,
			Body: ServerErrorMessage{Message: "failed to create subscription"},
		}
		c.SendJSON(r)
		return
	}

	c.subs[sub] = true

	r := Message{
		Type: MessageTypeSubAck,
		Body: m,
	}
	c.SendJSON(r)
}

func (c *Client) handleUnSubMessage(m *SubscriptionMessage) {
	var sub *Subscription
	for s := range c.subs {
		if s.sub.Subject == m.Channel {
			sub = s
			break
		}
	}

	if sub == nil {
		c.SendJSON(Message{
			Type: MessageTypeUnSubAck,
			Body: m,
		})
		return
	}

	delete(c.subs, sub)

	err := sub.Unsubscribe(c)
	if err != nil {
		log.Errorf("Error occurred while un-subscribing user from channel: %s", err)
		c.SendJSON(Message{
			Type: MessageTypeServerErr,
			Body: ServerErrorMessage{Message: "Error occurred while un-subscribing from channel: " + m.Channel},
		})
		return
	}
	c.SendJSON(Message{
		Type: MessageTypeUnSubAck,
		Body: m,
	})
}

func (c *Client) handleChatMessage(m *ChatMessage) {
	id, _ := uuid.NewV4()
	m.ID = id.String()
	found := false
	for s := range c.subs {
		if s.sub.Subject == m.Channel {
			found = true
			break
		}
	}
	if !found {
		c.SendJSON(Message{
			Type: MessageTypeForbiddenErr,
			Body: ServerErrorMessage{
				Message: "You are not authorized to publish to channel: " + m.Channel,
			},
		})
		return
	}

	b := bytes.Buffer{}
	err := gob.NewEncoder(&b).Encode(Message{
		Type: MessageTypeChat,
		Body: m,
	})
	if err != nil {
		log.Errorf("failed to serialize message to nats: %v", err)
		r := Message{
			Type: MessageTypeServerErr,
			Body: ServerErrorMessage{
				Message: "failed to process message",
			},
		}
		c.SendJSON(r)
		return
	}

	err = c.gnats.Publish(m.Channel, b.Bytes())
	if err != nil {
		log.Errorf("failed to send message to gnats server: %v", err)
		r := Message{
			Type: MessageTypeServerErr,
			Body: ServerErrorMessage{
				Message: "failed to process message",
			},
		}
		c.SendJSON(r)
		return
	}

	r := Message{
		Type: MessageTypeChatAck,
		Body: m,
	}
	c.SendJSON(r)
	return
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error while closing: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})

				var err error
				for s := range c.subs {
					if err = s.Unsubscribe(c); err != nil {
						log.Errorf("Got error while un-subscribing from channel: %v", err)
					}
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, gnats *Gnats, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(hub, gnats, conn)
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
