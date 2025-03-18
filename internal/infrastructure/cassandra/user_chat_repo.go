package cstorage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type UserChatRepo struct {
	instance *CassandraDB
	metric   appmetrics.Metrics
}

// CheckParticipants implements user.IUserRepo.
func (u *UserChatRepo) CheckParticipants(participants []string) ([]*user.Person, error) {
	l := len(participants)
	if l == 0 {
		return nil, errors.New("the list is empty")
	}
	var wg sync.WaitGroup
	errCh := make(chan error, l)
	personCh := make(chan *user.Person, l)
	wg.Add(l)
	for _, v := range participants {
		go func(id string) {
			defer wg.Done()
			var person _Person_cql
			err := person.readOne(u.instance, id)
			if err != nil {
				errCh <- err
				return
			}
			personCh <- &user.Person{
				Username:  person.Username,
				FirstName: person.Firstname,
				LastName:  person.Lastname,
				PersonId:  person.ID,
			}
		}(v)
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
	for v := range personCh {
		res = append(res, v)
	}
	return res, nil
}

// CreateNewChat implements user.IUserRepo.
func (u *UserChatRepo) CreateNewChat(uChat *user.UserChat) error {

	// 1- create chat row

	err := init_Chat(uChat).create(u.instance)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	errCh := make(chan error, 2)
	// 2- create chat by user
	go func() {
		defer wg.Done()
		err := new_ListOfUserByChat(uChat).upsert(u.instance)
		if err != nil {
			errCh <- err
		}
	}()

	// 3- create user by chat
	go func() {
		defer wg.Done()
		err := new__ListOfChatByUser(uChat).upsert(u.instance)
		if err != nil {
			errCh <- err
		}
	}()

	wg.Wait()
	close(errCh)

	// Check for any errors
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// GetUserChat implements user.IUserRepo.
func (u *UserChatRepo) GetUserChat(chatID string) (*user.UserChat, error) {
	chat, err := get_Chat(u.instance, chatID)
	if err != nil {
		return nil, err
	}
	return chat.ToUserChat(u.instance)
}

// GetUserHistory implements user.IUserRepo.
func (u *UserChatRepo) GetUserHistory(userID string) ([]*user.UserChat, error) {
	// 1- get chat by user
	var res []*user.UserChat
	list, err := get_ListOfChatByUser(u.instance, userID)
	if err != nil {
		return nil, err
	}
	l := len(*list)
	if l == 0 {
		return res, nil
	}
	var wg sync.WaitGroup
	errCh := make(chan error, l)
	cCh := make(chan *user.UserChat, l)
	wg.Add(l)
	for _, v := range *list {
		go func(cbu *_ChatByUser) {
			defer wg.Done()
			c, err := u.GetUserChat(cbu.chatID)
			if err != nil {
				errCh <- err
				return
			}
			if cbu.leaveAt != nil {
				c.LastMessageText = "Left the chat"
				c.LastMessageTime = *cbu.leaveAt
			}
			cCh <- c
		}(&v)
	}
	wg.Wait()
	close(errCh)
	close(cCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	for c := range cCh {
		res = append(res, c)
	}
	return res, nil
}

// SaveUserChat implements user.IUserRepo.
func (u *UserChatRepo) SaveUserChat(domain *user.UserChat) error {
	var list _ListOfUserByChat
	err := list.readMany(u.instance, domain.ChatID)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return errors.New("empty list")
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)
	go func(list _ListOfUserByChat) {
		defer wg.Done()
		err := list.edit(domain).upsert(u.instance)
		if err != nil {
			errCh <- err
		}
	}(list)
	go func(list _ListOfUserByChat) {
		defer wg.Done()
		cbuList, err := list.get_ListOfChatByUser(u.instance)
		if err != nil {
			fmt.Println("from 3")
			errCh <- err
			return
		}
		err = cbuList.edit(domain).upsert(u.instance)
		if err != nil {
			errCh <- err
		}
	}(list)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CassandraDB) NewUserChatRepo(metric appmetrics.Metrics) (*UserChatRepo, error) {
	return &UserChatRepo{
		instance: s,
		metric:   metric,
	}, nil
}
