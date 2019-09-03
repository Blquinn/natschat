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
	Type string      `validate:"required" json:"type"`
	Body interface{} `validate:"required" json:"body"`
}

func NewMessage(typ string, body interface{}) Message {
	return Message{
		Type: typ,
		Body: body,
	}
}

type SubscriptionMessage struct {
	Channel string `validate:"required" json:"channel"`
}

type ChatMessage struct {
	ID       string `json:"id"`
	ClientID string `validate:"required" json:"clientId"`
	Channel  string `validate:"required" json:"channel"`
	Content  string `validate:"required" json:"content"`
}

type ValidationErrorMessage struct {
	OriginalMessage string                    `json:"originalMessage"`
	Errors          []apierrs.ValidationError `json:"errors"`
}

type ServerErrorMessage struct {
	Message string `json:"message"`
}

func NewServerErrorMessage(message string) ServerErrorMessage {
	return ServerErrorMessage{
		Message: message,
	}
}
