package cstorage

import (
	"context"
	"sync"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
)

type _Chat struct {
	chatID    string
	chatName  string
	createdAt time.Time
	users     _ListOfUserByChat    // can be consider like a relation
	messages  _ListOfMessageByChat // can be consider like a relation
}

func init_Chat(uChat *user.UserChat) *_Chat {
	return &_Chat{
		chatID:    uChat.ChatID,
		chatName:  "", //TODO: add chat Name to UserChat
		createdAt: time.Now(),
	}
}

func get_Chat_row(instance *CassandraDB, chatID string) (*_Chat, error) {
	var row _Chat
	ctx := context.Background()
	err := instance.session.Query(`SELECT chat_id,chat_name,created_at FROM `+instance.cfg.Keyspace+`.chat WHERE chat_id=? LIMIT 1`, chatID).WithContext(ctx).Scan(&row.chatID, &row.chatName, &row.createdAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func get_Chat(instance *CassandraDB, chatID string) (*_Chat, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, 3)
	cCh := make(chan *_Chat, 1)
	ubcCh := make(chan _ListOfUserByChat, 1)
	mbcCh := make(chan _ListOfMessageByChat, 1)
	wg.Add(3)

	// 1- get chat
	go func() {
		defer wg.Done()
		chat, err := get_Chat_row(instance, chatID)
		if err != nil {
			errCh <- err
			return
		}
		cCh <- chat
	}()
	//2- get user by chat
	go func() {
		defer wg.Done()
		var val _ListOfUserByChat
		err := val.readMany(instance, chatID)
		if err != nil {
			errCh <- err
			return
		}
		ubcCh <- val
	}()
	//3- get last message by chat
	go func() {
		defer wg.Done()
		var val _ListOfMessageByChat
		err := val.getOnlyLast(instance, chatID)
		if err != nil {
			errCh <- err
			return
		}
		mbcCh <- val
	}()
	wg.Wait()
	close(errCh)
	close(cCh)
	close(ubcCh)
	close(mbcCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	chat := <-cCh
	chat.users = <-ubcCh
	chat.messages = <-mbcCh
	return chat, nil
}

func (chat *_Chat) ToUserChat(instance *CassandraDB) (*user.UserChat, error) {
	var uChat user.UserChat
	uChat.ChatID = chat.chatID
	l_of_m := len(chat.messages)
	if l_of_m > 1 {
		lastMessage := chat.messages[l_of_m-1]
		uChat.LastMessageText = lastMessage.text
		uChat.LastMessageTime = lastMessage.sentAt
	}
	// uChat.ChatName = chat.chatName
	// uChat.CreatedAt = chat.createdAt
	var err error
	uChat.Participants, err = chat.users.get_Persons(instance)
	if err != nil {
		return nil, err
	}

	return &uChat, nil
}

func (row *_Chat) create(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.chat (chat_id,chat_name,created_at) VALUES (?, ?, ?)`, row.chatID, row.chatName, row.createdAt).WithContext(ctx).Exec()
}

func readAll_Chat(instance *CassandraDB, chatID string) (*_Chat, error) {
	var wg sync.WaitGroup
	wg.Add(3)
	errCh := make(chan error, 3)
	chatCh := make(chan *_Chat, 1)
	usersCh := make(chan *_ListOfUserByChat, 1)
	messagesCh := make(chan *_ListOfMessageByChat, 1)

	go func() {
		defer wg.Done()
		chat, err := get_Chat_row(instance, chatID)
		if err != nil {
			errCh <- err
		}
		chatCh <- chat
	}()

	go func() {
		defer wg.Done()
		var users _ListOfUserByChat
		err := users.readMany(instance, chatID)
		if err != nil {
			errCh <- err
		}
		usersCh <- &users
	}()

	go func() {
		defer wg.Done()
		var messages _ListOfMessageByChat
		err := messages.readMany(instance, chatID)
		if err != nil {
			errCh <- err
		}
		messagesCh <- &messages
	}()
	wg.Wait()
	close(errCh)
	close(chatCh)
	close(usersCh)
	close(messagesCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	data := <-chatCh
	data.messages = *<-messagesCh
	data.users = *<-usersCh
	return data, nil
}
