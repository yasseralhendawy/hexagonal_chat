package cstorage

import (
	"context"
	"sync"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
)

type _ChatByUser struct {
	userID   string
	chatID   string
	chatName string
	isFav    bool
	joinAt   time.Time
	leaveAt  *time.Time
}

func (table *_ChatByUser) upsert(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.chat_by_user (user_id,chat_id,chat_name,is_fav,join_at,leave_at) VALUES (?, ?, ?, ?, ?, ?)`, table.userID, table.chatID, table.chatName, table.isFav, table.joinAt, table.leaveAt).WithContext(ctx).Exec()
}

func get_ChatByUser(instance *CassandraDB, chatID string, userID string) (*_ChatByUser, error) {
	var row _ChatByUser
	ctx := context.Background()
	err := instance.session.Query(`SELECT user_id,chat_id,chat_name,is_fav,join_at,leave_at FROM `+instance.cfg.Keyspace+`.chat_by_user WHERE user_id=? AND chat_id=? LIMIT 1`, userID, chatID).WithContext(ctx).Scan(&row.userID, &row.chatID, &row.chatName, &row.isFav, &row.joinAt, &row.leaveAt)
	if err != nil {
		return nil, err
	}
	return &row, nil

}

type _ListOfChatByUser []_ChatByUser

func new__ListOfChatByUser(uChat *user.UserChat) *_ListOfChatByUser {
	var list _ListOfChatByUser
	for _, participant := range uChat.Participants {
		row := _ChatByUser{
			userID:   participant.PersonId,
			chatID:   uChat.ChatID,
			chatName: "", // TODO: add chat name to UserChat
			isFav:    false,
			joinAt:   time.Now(),
		}
		list = append(list, row)
	}
	return &list
}
func (list *_ListOfChatByUser) edit(uChat *user.UserChat) *_ListOfChatByUser {
	// Create map of existing users for quick lookup
	existingUsers := make(map[string]bool)
	for _, chat := range *list {
		existingUsers[chat.userID] = true
	}

	// Add any new participants
	for _, participant := range uChat.Participants {
		if !existingUsers[participant.PersonId] {
			row := _ChatByUser{
				userID:   participant.PersonId,
				chatID:   uChat.ChatID,
				chatName: "", // TODO: add chat name to UserChat
				isFav:    false,
				joinAt:   time.Now(),
			}
			*list = append(*list, row)
		}
	}

	// Mark removed participants with leave time
	participantMap := make(map[string]bool)
	for _, p := range uChat.Participants {
		participantMap[p.PersonId] = true
	}

	now := time.Now()
	for i := range *list {
		if !participantMap[(*list)[i].userID] && (*list)[i].leaveAt == nil {
			(*list)[i].leaveAt = &now
		}
	}

	return list
}

func get_ListOfChatByUser(instance *CassandraDB, userID string) (*_ListOfChatByUser, error) {
	var list _ListOfChatByUser
	ctx := context.Background()
	scanner := instance.session.Query(`SELECT user_id,chat_id,chat_name,is_fav,join_at,leave_at FROM `+instance.cfg.Keyspace+`.chat_by_user WHERE user_id=? `, userID).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var row _ChatByUser
		err := scanner.Scan(&row.userID, &row.chatID, &row.chatName, &row.isFav, &row.joinAt, &row.leaveAt)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &list, nil
}

func (list *_ListOfChatByUser) upsert(instance *CassandraDB) error {
	var wg sync.WaitGroup
	l := len(*list)
	errCh := make(chan error, l)
	wg.Add(l)

	for _, chat := range *list {
		go func(c _ChatByUser) {
			defer wg.Done()
			err := c.upsert(instance)
			if err != nil {
				errCh <- err
			}
		}(chat)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// func (table *_ChatByUser) update(instance *CassandraDB) error {
// 	ctx := context.Background()
// 	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.chat_by_user SET chatname=? ,is_fav=? ,join_at=?,leave_at=? WHERE user_id=? AND chat_id=?`, table.chatName, table.isFav, table.joinAt, table.leaveAt, table.userID, table.chatID).WithContext(ctx).Exec()
// }
