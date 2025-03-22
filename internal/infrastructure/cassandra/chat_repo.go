package cstorage

import (
	"sync"

	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type ChatRepo struct {
	instance *CassandraDB
	metric   appmetrics.Metrics
}

func (s *CassandraDB) NewChatRepo(metric appmetrics.Metrics) (*ChatRepo, error) {
	return &ChatRepo{
		instance: s,
		metric:   metric,
	}, nil
}

// EditMessage implements chat.IChatRepo.
func (c *ChatRepo) EditMessage(message *chat.Message) error {
	_, err := get_Chat_row(c.instance, message.ChatID)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	errCh := make(chan error, 2)
	mbuCh := make(chan *_MessageByUser, 1)
	mbcCh := make(chan *_MessageByChat, 1)

	// check and get _MessageByUser is in the user messages
	go func() {
		defer wg.Done()
		mbu, err := get_MessageByUser(c.instance, message.MessageID, message.SenderID)
		if err != nil {
			errCh <- err
		}
		mbuCh <- mbu
	}()

	// check and _MessageByChat is in the user chats
	go func() {
		defer wg.Done()
		mbc, err := get_MessageByChat(c.instance, message.MessageID, message.ChatID, message.TimeToPost)
		if err != nil {
			errCh <- err
		}
		mbcCh <- mbc
	}()

	wg.Wait()
	close(errCh)
	close(mbcCh)
	close(mbuCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	mbu := <-mbuCh
	mbc := <-mbcCh
	wg.Add(2)
	errCh = make(chan error, 2)
	// update _MessageByUser
	go func() {
		defer wg.Done()
		err := mbu.edit(message).update(c.instance)
		if err != nil {
			errCh <- err
		}
	}()

	// update _MessageByChat
	go func() {
		defer wg.Done()
		err := mbc.edit(message).update(c.instance)
		if err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// GetChat implements chat.IChatRepo.
func (c *ChatRepo) GetChat(chatID string) (*chat.Chat, error) {
	value, err := readAll_Chat(c.instance, chatID)
	if err != nil {
		return nil, err
	}
	var res chat.Chat
	res.ChatID = value.chatID
	res.LastMessage = value.messages[0].sentAt
	for _, v := range value.users {
		res.ParticipantsIDs = append(res.ParticipantsIDs, v.userID)
	}
	for _, v := range value.messages {
		res.Messages = append(res.Messages, &chat.Message{
			MessageID:   v.messageID,
			SenderID:    v.userID,
			ChatID:      v.chatID,
			MessageText: v.text,
			TimeToPost:  v.sentAt,
			DeltedAt:    *v.deletedAt,
		})
	}
	return &res, nil
}

// SaveMessage implements chat.IChatRepo.
func (c *ChatRepo) SaveMessage(message *chat.Message) error {
	_, err := get_Chat_row(c.instance, message.ChatID)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	wg.Add(2)
	errCh := make(chan error, 2)
	go func() {
		defer wg.Done()
		err := new_MessageByChat(message).create(c.instance)
		if err != nil {
			errCh <- err
		}
	}()
	go func() {
		defer wg.Done()
		err := new_MessageByUser(message).create(c.instance)
		if err != nil {
			errCh <- err
		}
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
