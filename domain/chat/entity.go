package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	MessageID   string
	SenderID    string
	ChatID      string
	TimeToPost  time.Time
	MessageText string
	DeltedAt    time.Time
}

type MessageOpt func(*Message)

func (c *Chat) GetMessage(messageID string) (*Message, error) {
	for _, m := range c.Messages {
		if m.MessageID == messageID {
			return m, nil
		}
	}
	return nil, errors.New("message could not be found")
}

func NewMessage(senderID string, chatID string, text string) *Message {
	return &Message{
		MessageID:   uuid.New().String(),
		SenderID:    senderID,
		ChatID:      chatID,
		MessageText: text,
		TimeToPost:  time.Now(),
	}
}

func (m *Message) EditMessage(options ...MessageOpt) (*Message, error) {
	if len(options) == 0 {
		return m, errors.New("there is no options to operate")
	}
	for _, opt := range options {
		opt(m)
	}

	return m, nil
}

func MarkMessageAsDeleted() MessageOpt {
	return func(m *Message) {
		m.DeltedAt = time.Now()
	}
}

func EditText(text string) MessageOpt {
	return func(m *Message) {
		m.MessageText = text
	}
}
