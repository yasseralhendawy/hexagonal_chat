package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatOpt func(*UserChat) error

type UserChat struct {
	ChatID          string
	Participants    []*Person
	LastMessageText string
	LastMessageTime time.Time
}

func NewChat(participants []*Person) (*UserChat, error) {
	if len(participants) < 1 {
		return nil, errors.New("any chat should have at least one participants")
	}
	return &UserChat{
		ChatID:       uuid.New().String(),
		Participants: participants,
	}, nil
}

func (c *UserChat) EditChat(opts ...ChatOpt) (*UserChat, error) {
	if len(opts) == 0 {
		return c, errors.New("there is no options to operate")
	}
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func RemoveParticipant(userID string) ChatOpt {
	return func(uc *UserChat) error {
		for i, v := range uc.Participants {
			if v.PersonId == userID {
				uc.Participants = append(uc.Participants[:i], uc.Participants[i+1:]...)
				return nil
			}
		}
		return errors.New("can not find the participant")
	}
}

func AddParticipants(participants []*Person) ChatOpt {
	return func(uc *UserChat) error {
		if len(participants) == 0 {
			return errors.New("the participants list is empty")
		}
		participantsMap := make(map[string]bool)
		for _, v := range uc.Participants {
			participantsMap[v.PersonId] = true
		}
		for _, v := range participants {
			if participantsMap[v.PersonId] {
				return errors.New("there is participant already exist in this chat")
			}
			uc.Participants = append(uc.Participants, v)
			participantsMap[v.PersonId] = true
		}
		return nil
	}
}
