package main

import "natschat/utils/apierrs"

const (
	MessageTypeAuthAck         = "AUTHACK"
	MessageTypeSub             = "SUB"
	MessageTypeUnSub           = "UNSUB"
	MessageTypeChat            = "CHAT"
	MessageTypeSubAck          = "SUBACK"
	MessageTypeUnSubAck        = "UNSUBACK"
	MessageTypeChatAck         = "CHATACK"
	MessageTypeValidationErr   = "BAD"
	MessageTypeForbiddenErr    = "FORBIDDEN"
	MessageTypeUnauthorizedErr = "UNAUTHORIZED"
	MessageTypeServerErr       = "ERR"
)

var (
	TypeMap = map[string]func() interface{}{
		MessageTypeSub:   func() interface{} { return &SubscriptionMessage{} },
		MessageTypeUnSub: func() interface{} { return &SubscriptionMessage{} },
		MessageTypeChat:  func() interface{} { return &ChatMessage{} },
	}
)

type Message struct {
	Type string      `validate:"required"`
	Body interface{} `validate:"required"`
}

type SubscriptionMessage struct {
	Channel string `validate:"required"`
}

type ChatMessage struct {
	ID       string
	ClientID string `validate:"required"`
	Channel  string `validate:"required"`
	Content  string `validate:"required"`
}

type ValidationErrorMessage struct {
	OriginalMessage string
	Errors          []apierrs.ValidationError
}

type ServerErrorMessage struct {
	Message string
}
