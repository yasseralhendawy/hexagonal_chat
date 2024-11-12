package chat_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
)

type MockTestSuit struct {
	suite.Suite

	service *chat.ChatService
	repo    *chat.MockIChatRepo
}

func TestMockTestSuit(t *testing.T) {
	suite.Run(t, &MockTestSuit{})
}

func (mut *MockTestSuit) SetupTest() {
	mr := chat.MockIChatRepo{}

	mut.repo = &mr
	mut.service = &chat.ChatService{
		Storage: &mr,
		// Storage: chat.NewMockIChatRepo(),
	}

}

func (mut *MockTestSuit) TestGetChat() {
	var testCases = []struct {
		chatID         string
		expectedResult *chat.Chat
		expectedError  error
	}{
		{chatID: "cid", expectedResult: nil, expectedError: errors.New("not found")},
		{chatID: "cid", expectedResult: &chat.Chat{ChatID: "cid"}, expectedError: nil},
	}

	for _, c := range testCases {
		mut.repo.EXPECT().GetChat(c.chatID).Return(c.expectedResult, c.expectedError).Times(1)
		res, err := mut.service.GetChat(c.chatID)
		mut.Assert().Equal(c.expectedError, err)
		mut.Assert().Equal(c.expectedResult, res)
		if c.expectedResult != nil {
			mut.Assert().Equal(c.chatID, res.ChatID)
		}
	}
}

var chatTemp = chat.Chat{
	ChatID:          "id",
	ParticipantsIDs: []string{"id1,id2"},
	Messages: []*chat.Message{
		{
			MessageID:   "mid1",
			SenderID:    "senderID",
			ChatID:      "id",
			MessageText: "text",
		},
	},
	LastMessage: time.Now().Add(-1 * time.Minute),
}

func (mut *MockTestSuit) TestAddMessage() {

	var testCases = []struct {
		// chat       *chat.Chat // u can uncomment and add the chat
		chatID     string
		getChaterr error
		saveErr    error
	}{
		{chatID: "", getChaterr: errors.New("can not get the chat"), saveErr: nil},
		{chatID: "id", getChaterr: nil, saveErr: errors.New("save error")},
		{chatID: chatTemp.ChatID, getChaterr: nil, saveErr: nil},
	}

	for _, c := range testCases {
		chat := chatTemp
		mut.repo.EXPECT().GetChat(c.chatID).Return(&chat, c.getChaterr).Times(1)
		if c.getChaterr == nil {
			mut.repo.EXPECT().SaveMessage(mock.Anything).Return(c.saveErr).Times(1)
		}
		res, err := mut.service.AddMesage(c.chatID, mock.Anything, mock.Anything)
		if err != nil {
			if c.getChaterr != nil {
				mut.Assert().Equal(c.getChaterr, err)
			} else {
				mut.Assert().Equal(c.saveErr, err)
			}
		} else {
			mut.Assert().Equal(c.chatID, res.ChatID)
			mut.Assert().NotEqual(chatTemp, res)
			mut.Assert().Equal(2, len(res.Messages)) // as we use only one temp of chat which has only one message so it's expected to be the result 2
			mut.Assert().NotEqual(len(chatTemp.Messages), len(res.Messages))
		}
	}
}

func (mut *MockTestSuit) TestEditChatMessage() {
	var testCases = []struct {
		opt         chat.MessageOpt
		messageID   string
		userID      string
		chatID      string
		getChaterr  error
		editErr     error
		expectedErr bool
	}{
		{opt: chat.EditText("text"), messageID: "mid1", userID: "user", chatID: "id", getChaterr: nil, editErr: nil, expectedErr: true}, // user id is different
		{opt: chat.EditText("text"), messageID: "mid1", userID: "senderID", chatID: "id", getChaterr: nil, editErr: nil, expectedErr: false},
		{opt: chat.MarkMessageAsDeleted(), messageID: "mid1", userID: "senderID", chatID: "id", getChaterr: nil, editErr: nil, expectedErr: false},
	}

	for _, c := range testCases {
		chat := chatTemp
		mut.repo.EXPECT().GetChat(c.chatID).Return(&chat, c.getChaterr).Times(1)
		if c.getChaterr == nil {
			mut.repo.EXPECT().EditMessage(mock.Anything).Return(c.editErr).Times(1)
		}
		res, err := mut.service.EditChatMessage(c.userID, c.chatID, c.messageID, c.opt)
		if c.expectedErr {
			mut.Assert().Error(err)
			if c.getChaterr != nil {
				mut.Assert().Equal(c.getChaterr, err)
			}
			if c.editErr != nil {
				mut.Assert().Equal(c.editErr, err)
			}
		} else {
			mut.Require().NotNil(res)
			mut.Assert().Equal(c.chatID, res.ChatID)
			mut.Assert().NotEqual(chatTemp, res)
			mut.Assert().Equal(len(chatTemp.Messages), len(res.Messages))
		}
	}
}
