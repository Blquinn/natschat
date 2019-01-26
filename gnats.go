package main

import (
	"bytes"
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"sync"

	nats "github.com/nats-io/go-nats"
)

// NATS client creates a new subscription on the nats server

type Subscription struct {
	mu sync.Mutex

	gnats   *Gnats
	sub     *nats.Subscription
	clients map[*Client]bool
}

func newSubscription(gnats *Gnats, sub *nats.Subscription, client *Client) *Subscription {
	return &Subscription{
		gnats:   gnats,
		sub:     sub,
		clients: map[*Client]bool{client: true},
	}
}

func (s *Subscription) Unsubscribe(c *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.clients[c] {
		log.Debugln("Client", c.conn.RemoteAddr(), "unsubscribed from non-member channel:", s.sub.Subject)
		return nil
	}

	delete(s.clients, c)

	if len(s.clients) == 0 {
		err := s.sub.Unsubscribe()
		s.gnats.mu.Lock()
		delete(s.gnats.subs, s.sub.Subject)
		s.gnats.mu.Unlock()
		if err != nil {
			log.Errorf("Error occurred while un-subscribing from nats subject: %v", err)
			return err
		}
	}
	log.Debugf("Successfully un-subscribed client %s from channel %s", c.conn.RemoteAddr(), s.sub.Subject)
	return nil
}

type Gnats struct {
	mu sync.RWMutex

	conn *nats.Conn
	subs map[string]*Subscription
}

func newGnats(c *nats.Conn) *Gnats {
	return &Gnats{
		conn: c,
		subs: make(map[string]*Subscription),
	}
}

func (g *Gnats) Publish(ch string, msg []byte) error {
	return g.conn.Publish(ch, msg)
}

func (g *Gnats) Subscribe(ch string, client *Client) (*Subscription, error) {

	g.mu.RLock()
	sub, f := g.subs[ch]
	g.mu.RUnlock()

	if f {
		sub.mu.Lock()
		sub.clients[client] = true
		log.Debugf("Successfully subscribed client %s to channel %s", client.conn.RemoteAddr(), sub.sub.Subject)
		sub.mu.Unlock()
		return sub, nil
	}

	s, err := g.conn.Subscribe(ch, g.subscriptionHandler)
	if err != nil {
		return nil, err
	}
	newSub := newSubscription(g, s, client)
	g.mu.Lock()
	g.subs[ch] = newSub
	g.mu.Unlock()
	log.Debugf("Successfully subscribed client %s to channel %s", client.conn.RemoteAddr(), newSub.sub.Subject)
	return newSub, nil
}

func (g *Gnats) subscriptionHandler(msg *nats.Msg) {
	m := Message{}
	err := gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&m)
	if err != nil {
		log.Errorf("failed to decode gob message from gnats: %v", err)
		return
	}

	if m.Type == MessageTypeChat {
		g.mu.RLock()
		sub, f := g.subs[msg.Subject]
		g.mu.RUnlock()
		if f {
			for c := range sub.clients {
				c.SendJSON(m)
			}
		}
		return
	}

	log.Warnf("got unhandled msg from nats: %v", m)
}
