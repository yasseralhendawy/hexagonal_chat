package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ChatID          string
	ParticipantsIDs []string
	Messages        []*Message
	LastMessage     time.Time
}

type ChatOpt func(*Chat) error

func NewChat(participants []string) (*Chat, error) {
	if len(participants) < 2 {
		return nil, errors.New("any chat should have at least two participants")
	}
	return &Chat{
		ChatID:          uuid.New().String(),
		ParticipantsIDs: participants,
	}, nil
}

func (c *Chat) EditChat(options ...ChatOpt) (*Chat, error) {
	if len(options) == 0 {
		return c, errors.New("there is no options to operate")
	}
	for _, opt := range options {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func AddMessage(message *Message) ChatOpt {
	return func(c *Chat) error {
		if c.ChatID != message.ChatID {
			return errors.New("the chat id is diffent")
		}
		for _, v := range c.Messages {
			if v.MessageID == message.MessageID {
				return errors.New("this message is already exist")
			}
		}
		c.Messages = append(c.Messages, message)
		c.LastMessage = message.TimeToPost
		return nil
	}
}

func EditMessage(message *Message) ChatOpt {
	return func(c *Chat) error {
		if c.ChatID != message.ChatID {
			return errors.New("the chat id is diffent")
		}
		for i, v := range c.Messages {
			if v.MessageID == message.MessageID {
				c.Messages[i] = message
				return nil
			}
		}
		return errors.New("this message can not be found")
	}
}