package chat

import (
	"errors"
	"slices"
	"time"
)

type Chat struct {
	ChatID          string
	ParticipantsIDs []string
	Messages        []*Message
	LastMessage     time.Time
}

type ChatOpt func(*Chat) error

func (c *Chat) GetMessage(messageID string) (*Message, error) {
	for _, m := range c.Messages {
		if m.MessageID == messageID {
			return m, nil
		}
	}
	return nil, errors.New("message could not be found")
}

func (c *Chat) EditChat(options ...ChatOpt) (*Chat, error) {
	if len(options) == 0 {
		return nil, errors.New("there is no options to operate")
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
		found := slices.Contains(c.ParticipantsIDs, message.SenderID)
		if !found {
			return errors.New("user is not in the chat")
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
