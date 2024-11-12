package user

import (
	"errors"
)

type IUserRepo interface {
	CreateNewChat(*UserChat) error
	GetUserChat(string) (*UserChat, error)
	SaveUserChat(*UserChat) error
	// GetPerson(string) (*Person, error)
	CheckParticipants(participants []string) ([]*Person, error)
	GetUserHistory(string) ([]*UserChat, error)
}

type Service struct {
	Storage IUserRepo
}

func (s *Service) CreatNewChat(participants []string, senderID string, text string) (*UserChat, error) {
	if len(participants) < 1 {
		return nil, errors.New("any chat should have at least one participants")
	}
	//lets add a rule that all participants should be exist
	persons, err := s.Storage.CheckParticipants(participants)
	if err != nil {
		return nil, err
	}
	c, err := NewChat(persons)
	if err != nil {
		return nil, err
	}
	err = s.Storage.CreateNewChat(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetUserHistory(userID string) ([]*UserChat, error) {
	return s.Storage.GetUserHistory(userID)
}

// func to leave chat
func (s *Service) LeaveChat(userID string, chatID string) error {
	_, err := s.EditChat(chatID, RemoveParticipant(userID))
	return err
}

func (s *Service) AddParticipants(chatID string, participants []string) (*UserChat, error) {
	persons, err := s.Storage.CheckParticipants(participants)
	if err != nil {
		return nil, err
	}
	return s.EditChat(chatID, AddParticipants(persons))
}

func (s *Service) EditChat(chatID string, opt ...ChatOpt) (*UserChat, error) {
	if len(opt) == 0 {
		return nil, errors.New("there is no options to operate")
	}
	c, err := s.Storage.GetUserChat(chatID)
	if err != nil {
		return nil, err
	}
	c, err = c.EditChat(opt...)
	if err != nil {
		return nil, err
	}
	err = s.Storage.SaveUserChat(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// func (s *Service) checkParticipants(participants []string) ([]*Person, error) {
// 	var g errgroup.Group
// 	ch := make(chan *Person, len(participants))

// 	persons := make([]*Person, len(participants))
// 	for _, v := range participants {
// 		g.Go(func() error {
// 			p, err := s.Storage.GetPerson(v)
// 			if err != nil {
// 				return err
// 			}
// 			ch <- p
// 			return nil
// 		})
// 	}
// 	err := g.Wait()
// 	close(ch)

// 	if err != nil {
// 		return nil, err
// 	}
// 	for v := range ch {
// 		persons = append(persons, v)
// 	}
// 	return persons, nil
// }
