package cstorage

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
)

type _UserByChat struct {
	chatID  string
	userID  string
	isAdmin bool
	joinAt  time.Time
	leaveAt *time.Time
}

func (table *_UserByChat) upsert(instance *CassandraDB) error {
	ctx := context.Background()
	return instance.session.Query(`INSERT INTO `+instance.cfg.Keyspace+`.user_by_chat (chat_id,user_id,is_admin,join_at,leave_at) VALUES (?, ?, ?, ?, ?)`, table.chatID, table.userID, table.isAdmin, table.joinAt, table.leaveAt).WithContext(ctx).Exec()
}

// func (table *_UserByChat) check(instance *CassandraDB, chatID string, userID string) (bool, error) {
// 	ctx := context.Background()
// 	scanner := instance.session.Query(`SELECT chat_id,user_id,is_admin,join_at,leave_at FROM `+instance.cfg.Keyspace+`.user_by_chat WHERE chat_id=? AND user_id=? LIMIT 1`, chatID, userID).WithContext(ctx).Iter().Scanner()
// 	return scanner.Next(), scanner.Err()

// }

type _ListOfUserByChat []_UserByChat

func new_ListOfUserByChat(uChat *user.UserChat) *_ListOfUserByChat {
	var list _ListOfUserByChat
	for _, participant := range uChat.Participants {
		row := _UserByChat{
			chatID:  uChat.ChatID,
			userID:  participant.PersonId,
			isAdmin: false,
			joinAt:  time.Now(),
		}
		list = append(list, row)
	}
	return &list
}

func (list *_ListOfUserByChat) get_Persons(instance *CassandraDB) ([]*user.Person, error) {
	var wg sync.WaitGroup
	n := len(*list)
	errCh := make(chan error, n)
	personCh := make(chan *user.Person, n)
	wg.Add(n)
	for _, v := range *list {
		go func(userID string) {
			defer wg.Done()
			var p _Person_cql
			err := p.readOne(instance, userID)
			if err != nil {
				errCh <- err
				return
			}
			personCh <- &user.Person{
				FirstName: p.Firstname,
				LastName:  p.Lastname,
				PersonId:  p.ID,
				Username:  p.Username,
			}
		}(v.userID)
	}

	wg.Wait()
	close(errCh)
	close(personCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	var res []*user.Person
	for p := range personCh {
		res = append(res, p)
	}
	return res, nil
}

func (list *_ListOfUserByChat) get_ListOfChatByUser(instance *CassandraDB) (*_ListOfChatByUser, error) {
	l := len(*list)
	if l == 0 {
		return nil, errors.New("empty list")
	}
	var wg sync.WaitGroup
	errCh := make(chan error, l)
	cbuCh := make(chan _ChatByUser, l)
	wg.Add(l)
	for _, v := range *list {
		go func(chatID string, userID string) {
			defer wg.Done()
			c, err := get_ChatByUser(instance, chatID, userID)
			if err != nil {
				errCh <- err
				return
			}
			cbuCh <- *c
		}(v.chatID, v.userID)
	}
	wg.Wait()
	close(errCh)
	close(cbuCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	var res _ListOfChatByUser
	for c := range cbuCh {
		res = append(res, c)
	}
	return &res, nil
}

func (list *_ListOfUserByChat) edit(uChat *user.UserChat) *_ListOfUserByChat {
	// Create map of existing users for quick lookup
	existingUsers := make(map[string]bool)
	for _, user := range *list {
		existingUsers[user.userID] = true
	}

	// Add any new participants
	for _, participant := range uChat.Participants {
		if !existingUsers[participant.PersonId] {
			row := _UserByChat{
				chatID:  uChat.ChatID,
				userID:  participant.PersonId,
				isAdmin: false, // todo
				joinAt:  time.Now(),
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

func (list *_ListOfUserByChat) readMany(instance *CassandraDB, chatID string) error {
	ctx := context.Background()
	scanner := instance.session.Query(`SELECT chat_id,user_id,is_admin,join_at,leave_at FROM `+instance.cfg.Keyspace+`.user_by_chat WHERE chat_id=?`, chatID).WithContext(ctx).Iter().Scanner()
	for scanner.Next() {
		var row _UserByChat
		err := scanner.Scan(&row.chatID, &row.userID, &row.isAdmin, &row.joinAt, &row.leaveAt)
		if err != nil {
			return err
		}
		*list = append(*list, row)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (list *_ListOfUserByChat) upsert(instance *CassandraDB) error {
	l := len(*list)
	var wg sync.WaitGroup
	errCh := make(chan error, l)
	wg.Add(l)
	for _, user := range *list {
		go func(u _UserByChat) {
			defer wg.Done()
			err := u.upsert(instance)
			if err != nil {
				errCh <- err
			}
		}(user)
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

// func (table *_UserByChat) update(instance *CassandraDB) error {
// 	ctx := context.Background()
// 	return instance.session.Query(`UPDATE `+instance.cfg.Keyspace+`.user_by_chat SET is_admin=? ,join_at=?,leave_at=? WHERE chat_id=? AND user_id=?`, table.isAdmin, table.joinAt, table.leaveAt, table.chatID, table.userID).WithContext(ctx).Exec()
// }

// func (table *_UserByChat) readOne(instance *CassandraDB, chatID string, userID string) error {
// 	ctx := context.Background()
// 	err := instance.session.Query(`SELECT chat_id,user_id,is_admin,join_at,leave_at FROM `+instance.cfg.Keyspace+`.user_by_chat WHERE chat_id=? AND user_id=? LIMIT 1`, chatID, userID).WithContext(ctx).Scan(&table.chatID, &table.userID, &table.isAdmin, &table.joinAt, &table.leaveAt)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
