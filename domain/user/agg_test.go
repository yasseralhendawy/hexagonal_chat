package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
)

func TestNewChat(t *testing.T) {
	var testCases = []struct {
		participants []*user.Person
		expectedErr  bool
	}{
		{expectedErr: true},
		{participants: []*user.Person{
			{Username: "username", PersonId: "id", LastName: "last", FirstName: "first"},
		}, expectedErr: false},
	}
	assert := assert.New(t)

	for _, c := range testCases {

		res, err := user.NewChat(c.participants)
		if c.expectedErr {
			assert.Error(err)
			assert.Nil(res)
		} else {
			assert.Nil(err)
			assert.NotNil(res)
		}
	}
}

var tempChat = user.UserChat{
	ChatID:          "chatID",
	LastMessageText: "hi",
	LastMessageTime: time.Now().Add(-1 * time.Minute),
	Participants: []*user.Person{
		{Username: "username", PersonId: "id", LastName: "last", FirstName: "first"},
	},
}

func TestEditChat(t *testing.T) {
	var testCases = []struct {
		opts        []user.ChatOpt
		expectedErr bool
	}{
		{expectedErr: true}, // as there is no options
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{})}, expectedErr: true}, // as the participants is empty
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id", LastName: "last", FirstName: "first"},
		})}, expectedErr: true}, //there is participant already exist in this chat
		{opts: []user.ChatOpt{user.RemoveParticipant("id2")}, expectedErr: true}, // user is not found

		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		})}, expectedErr: false},
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		}), user.RemoveParticipant("id")}, expectedErr: false},
	}

	assert := assert.New(t)

	for _, c := range testCases {
		tc := tempChat
		res, err := tc.EditChat(c.opts...)
		if c.expectedErr {
			assert.Error(err)
			assert.Nil(res)
		} else {
			assert.NotEqual(tempChat, res)
			assert.Nil(err)
		}
	}
}
