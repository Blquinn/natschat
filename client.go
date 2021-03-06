package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"natschat/components/chat"
	"natschat/utils/apierrs"
	"natschat/utils/auth"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/go-playground/validator.v8"

	"github.com/gorilla/websocket"
)

func init() {
	gob.Register(ChatMessage{})
	gob.Register(chat.ChatMessageDTO{})
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

	user *auth.JWTUser

	cs *chat.Service

	validate *validator.Validate

	// Buffered channel of outbound messages.
	send chan []byte
}

func newClient(hub *Hub, gnats *Gnats, conn *websocket.Conn, cs *chat.Service, user *auth.JWTUser) *Client {
	return &Client{
		hub:      hub,
		gnats:    gnats,
		subs:     make(map[*Subscription]bool),
		conn:     conn,
		cs:       cs,
		user:     user,
		validate: validator.New(&validator.Config{TagName: "validate"}),
		send:     make(chan []byte, 100),
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

	err = c.validate.Struct(msg)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			log.Debugf("error while validating struct: %v", err)
			c.send <- []byte("internal server error")
			return
		}

		ves := apierrs.FormatValidationErrors(errs)

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
		c.send <- errMsg
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
	found := false
	for s := range c.subs {
		if s.sub.Subject == m.Channel {
			found = true
			break
		}
	}
	if !found {
		msg := "You are not authorized to publish to channel: " + m.Channel
		c.SendJSON(Message{Type: MessageTypeForbiddenErr, Body: ServerErrorMessage{Message: msg}})
		return
	}

	chunks := strings.Split(m.Channel, ".")
	if len(chunks) < 3 {
		c.SendJSON(Message{Type: MessageTypeServerErr, Body: ServerErrorMessage{Message: "Sever error occurred"}})
		log.Errorf("Unable to parse channel: %s", m.Channel)
		return
	}

	dbMsg, err := c.cs.SaveChatMessage(m.Content, chunks[2], c.user)
	if err != nil {
		if err.IsPublic {
			c.SendJSON(Message{Type: MessageTypeValidationErr, Body: ServerErrorMessage{Message: err.Message}})
		} else {
			c.SendJSON(Message{Type: MessageTypeServerErr, Body: ServerErrorMessage{Message: "Sever error occurred"}})
			log.Errorf("Error occurred while saving chat message %v", err)
		}
		return
	}
	m.ID = dbMsg.ID

	b := bytes.Buffer{}
	msg := Message{Type: MessageTypeChat, Body: dbMsg}
	if err := gob.NewEncoder(&b).Encode(msg); err != nil {
		log.Errorf("failed to serialize message to nats: %v", err)
		r := NewMessage(MessageTypeServerErr, NewServerErrorMessage("failed to process message"))
		c.SendJSON(r)
		return
	}

	if err := c.gnats.Publish(m.Channel, b.Bytes()); err != nil {
		log.Errorf("failed to send message to gnats server: %v", err)
		r := NewMessage(MessageTypeServerErr, NewServerErrorMessage("failed to process message"))
		c.SendJSON(r)
		return
	}
	log.Debugf("Successfully sent chat message to gnatsd channel %s", m.Channel)

	c.SendJSON(NewMessage(MessageTypeChatAck, m))
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			log.Errorf("got err while closing websocket connection: %v", err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Errorf("got err while setting read deadline on ws: %v", err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Warnf("got err while setting read deadline in ponghandler: %v", err)
		}
		return nil
	})
	for {
		t, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error while closing: %v", err)
			}
			break
		}

		if t == websocket.BinaryMessage {
			log.Warnf("Client %s sent a binary message. Closing socket.", c.conn.RemoteAddr())
			c.SendJSON(NewMessage(MessageTypeValidationErr, NewServerErrorMessage("Binary messages not accepted.")))
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
		if err := c.conn.Close(); err != nil {
			log.Errorf("got err while closing websocket connection: %v", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Errorf("got err wile setting write deadline for websocket msg: %v", err)
			}
			if !ok {
				log.Debugf("User send channel closed, closing websocket")
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Errorf("got error while writing close message: %v", err)
				}

				var err error
				for s := range c.subs {
					if err = s.Unsubscribe(c); err != nil {
						log.Errorf("Got error while un-subscribing from channel: %v", err)
					}
				}
				return
			}

			if err := sendWebsocketMessage(c.conn, message, websocket.TextMessage); err != nil {
				log.Errorf("got err while sending websocket message: %v", err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if err := sendWebsocketMessage(c.conn, <-c.send, websocket.TextMessage); err != nil {
					log.Errorf("got err while sending websocket msg: %v", err)
				}
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Errorf("got err while setting write deadline on websocket: %v", err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("got err while sending ping message over websocket: %v", err)
				return
			}
		}
	}
}

func sendWebsocketMessage(c *websocket.Conn, msg []byte, msgType int) error {
	w, err := c.NextWriter(msgType)
	if err != nil {
		return err
	}

	if _, err := w.Write(msg); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
