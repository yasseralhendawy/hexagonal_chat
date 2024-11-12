package chat

import (
	"errors"
)

type IChatRepo interface {
	GetChat(string) (*Chat, error)
	SaveMessage(*Message) error
	EditMessage(*Message) error
}

type ChatService struct {
	Storage IChatRepo
}

func (s *ChatService) GetChat(chatID string) (*Chat, error) {
	return s.Storage.GetChat(chatID)
}

func (s *ChatService) AddMesage(chatID string, senderID string, text string) (*Chat, error) {
	c, err := s.GetChat(chatID)
	if err != nil {
		return nil, err
	}
	message := NewMessage(senderID, chatID, text)
	c, err = c.EditChat(AddMessage(message))
	if err != nil {
		return nil, err
	}
	err = s.Storage.SaveMessage(message)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ChatService) EditMessageText(userID string, chatID string, messageID string, text string) (*Chat, error) {
	return s.EditChatMessage(userID, chatID, messageID, EditText(text))
}

func (s *ChatService) DeleteMessage(userID string, chatID string, messageID string) (*Chat, error) {
	return s.EditChatMessage(userID, chatID, messageID, MarkMessageAsDeleted())
}

func (s *ChatService) EditChatMessage(userID string, chatID string, messageID string, opt MessageOpt) (*Chat, error) {
	c, err := s.GetChat(chatID)
	if err != nil {
		return nil, err
	}
	message, err := c.GetMessage(messageID)
	if err != nil {
		return nil, err
	}
	if message.SenderID != userID {
		return nil, errors.New("this message does not belong to the user")
	}
	message, err = message.EditMessage(opt)
	if err != nil {
		return nil, err
	}
	c, err = c.EditChat(EditMessage(message))
	if err != nil {
		return nil, err
	}
	err = s.Storage.EditMessage(message)
	if err != nil {
		return nil, err
	}
	return c, nil
}
