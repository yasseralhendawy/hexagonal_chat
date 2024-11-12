package user_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
)

type MockTestSuit struct {
	suite.Suite
	repo    *user.MockIUserRepo
	service *user.Service
}

func TestMockTestSuit(t *testing.T) {
	suite.Run(t, &MockTestSuit{})
}

func (mut *MockTestSuit) SetupTest() {
	mut.repo = &user.MockIUserRepo{}
	mut.service = &user.Service{Storage: mut.repo}
}

func (mut *MockTestSuit) TestCreateNewChat() {
	var testCases = []struct {
		participants []string
		checkErr     error
		saveErr      error
		expectedErr  bool
	}{
		{participants: []string{}, expectedErr: true},
		{participants: []string{"id1", "id2"}, expectedErr: false, checkErr: nil, saveErr: nil},
	}

	for _, c := range testCases {
		if len(c.participants) > 1 {
			var persons []*user.Person
			if c.checkErr == nil {
				for _, p := range c.participants {
					persons = append(persons, &user.Person{Username: mock.Anything, PersonId: p})
				}
			}
			mut.repo.EXPECT().CheckParticipants(c.participants).Return(persons, c.checkErr).Once()
			if c.checkErr == nil {
				mut.repo.On("CreateNewChat", mock.Anything).Return(c.saveErr).Once()
			}
		}
		res, err := mut.service.CreatNewChat(c.participants, mock.Anything, mock.Anything)
		if c.expectedErr {
			mut.Assert().Error(err)
		} else {
			mut.NotNil(res)
		}
	}
}

func (mut *MockTestSuit) TestEditChat() {

	var testCases = []struct {
		opts      []user.ChatOpt
		getErr    error
		domainErr bool
		saveErr   error
	}{
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		})}, domainErr: false, getErr: errors.New("fail"), saveErr: nil},

		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{})}, domainErr: true, getErr: nil}, // as the participants is empty
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id", LastName: "last", FirstName: "first"},
		})}, domainErr: true, getErr: nil}, //there is participant already exist in this chat
		{opts: []user.ChatOpt{user.RemoveParticipant("id2")}, domainErr: true, getErr: nil}, // user is not found

		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		})}, domainErr: false, getErr: nil, saveErr: nil},
		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		}), user.RemoveParticipant("id")}, domainErr: false, getErr: nil, saveErr: nil},

		{opts: []user.ChatOpt{user.AddParticipants([]*user.Person{
			{Username: "username", PersonId: "id2", LastName: "last", FirstName: "first"},
		})}, domainErr: false, getErr: nil, saveErr: errors.New("fail")},
	}

	for _, c := range testCases {
		ch := tempChat
		mut.repo.EXPECT().GetUserChat(mock.Anything).Return(&ch, c.getErr).Once()
		if !c.domainErr && c.getErr == nil {
			mut.repo.On("SaveUserChat", mock.Anything).Return(c.saveErr).Once()
		}

		res, err := mut.service.EditChat(ch.ChatID, c.opts...)
		if c.domainErr || c.getErr != nil || c.saveErr != nil {
			mut.Assert().Error(err)
			mut.Assert().Nil(res)
		} else {
			mut.Assert().Nil(err)
			mut.Assert().NotNil(res)
			mut.Assert().Equal(ch.ChatID, res.ChatID)
		}
	}
}
