package cstorage_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	"github.com/yasseralhendawy/hexagonal_chat/domain/chat"
	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type _ChatRepoSuit struct {
	suite.Suite
	session      *cstorage.CassandraDB
	repo         *cstorage.ChatRepo
	authRepo     *cstorage.AuthRepo
	userChatRepo *cstorage.UserChatRepo
}

func (uts *_ChatRepoSuit) SetupSuite() {
	var err error
	uts.session, err = cstorage.NewCassandraSession(cfg)
	uts.Require().Nil(err)
	metric := &appmetrics.MockMetrics{}
	metric.EXPECT().DBCallsWithLabelValues(mock.Anything, mock.Anything, mock.Anything).Return()
	uts.repo, err = uts.session.NewChatRepo(metric)
	uts.Require().Nil(err)
	uts.userChatRepo, err = uts.session.NewUserChatRepo(metric)
	uts.Require().Nil(err)
	uts.authRepo, err = uts.session.NewAuthRepo(metric)
	uts.Require().Nil(err)
}

var tnow time.Time

func (uts *_ChatRepoSuit) BeforeTest(suiteName, testName string) {
	fmt.Printf("Before running %s.%s \n", suiteName, testName)
	tnow = time.Now()
	switch testName {
	case "TestSaveMessage":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "sm_user", Email: "sm_user@mail.com", Username: "sm_user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.authRepo.CreateNewUser(&auth.User{UserId: "sm_user2", Email: "sm_user2@mail.com", Username: "sm_user2", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.userChatRepo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "sm_user", Username: "sm_user"},
			{PersonId: "sm_user2", Username: "sm_user2"},
		},
			ChatID: "sm_chat",
		})
		uts.Require().Nil(err)
	case "TestGetChat":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "gc_user", Email: "gc_user@mail.com", Username: "gc_user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.authRepo.CreateNewUser(&auth.User{UserId: "gc_user2", Email: "gc_user2@mail.com", Username: "gc_user2", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.userChatRepo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "gc_user", Username: "gc_user"},
			{PersonId: "gc_user2", Username: "gc_user2"},
		},
			ChatID: "gc_chat",
		})
		uts.Require().Nil(err)
		err = uts.repo.SaveMessage(chat.NewMessage("gc_user", "gc_chat", "hi"))
		uts.Require().Nil(err)
		err = uts.repo.SaveMessage(chat.NewMessage("gc_user2", "gc_chat", "hi back"))
		uts.Require().Nil(err)
	case "TestEditMessage":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "em_user", Email: "em_user@mail.com", Username: "em_user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.authRepo.CreateNewUser(&auth.User{UserId: "em_user2", Email: "em_user2@mail.com", Username: "em_user2", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.userChatRepo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "em_user", Username: "em_user"},
			{PersonId: "em_user2", Username: "em_user2"},
		},
			ChatID: "em_chat",
		})
		uts.Require().Nil(err)
		err = uts.repo.SaveMessage(&chat.Message{
			MessageID:   "em_message",
			SenderID:    "em_user",
			ChatID:      "em_chat",
			MessageText: "text",
			TimeToPost:  tnow,
		})
		uts.Require().Nil(err)
		err = uts.repo.SaveMessage(&chat.Message{
			MessageID:   "em_message2",
			SenderID:    "em_user2",
			ChatID:      "em_chat",
			MessageText: "text 2",
			TimeToPost:  tnow,
		})
		uts.Require().Nil(err)
	default:
	}
}

func (uts *_ChatRepoSuit) TearDownSuite() {
	err := uts.session.DropTables()
	uts.Require().Nil(err)
}

func Test_ChatRepoTestSuit(t *testing.T) {
	suite.Run(t, &_ChatRepoSuit{})
}

func (uts *_ChatRepoSuit) TestSaveMessage() {
	var testcase = []struct {
		message   *chat.Message
		expectErr bool
	}{
		{
			message:   chat.NewMessage("sender", "sm_chat", "some text"),
			expectErr: false, // we assume the domain handled it before
		},
		{
			message:   chat.NewMessage("sm_user", "chat_not_exist", "some text"),
			expectErr: true, // as there is no chat id with the id "chat_not_exist"
		},
		{
			message:   chat.NewMessage("sm_user", "sm_chat", "some text"),
			expectErr: false,
		},
		{
			message:   chat.NewMessage("sm_user2", "sm_chat", "some text"),
			expectErr: false,
		},
	}

	for _, v := range testcase {
		err := uts.repo.SaveMessage(v.message)
		if v.expectErr {
			uts.Assert().Error(err, *v.message)
		} else {
			uts.Assert().Nil(err, *v.message)
		}
	}
}

func (uts *_ChatRepoSuit) TestGetChat() {
	var testCases = []struct {
		chatId    string
		expectErr bool
	}{
		{chatId: "not exist", expectErr: true},
		{chatId: "gc_chat", expectErr: false},
	}
	for _, v := range testCases {
		res, err := uts.repo.GetChat(v.chatId)
		if v.expectErr {
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
			fmt.Println(" res is ", res)
		}
	}
}

func (uts *_ChatRepoSuit) TestEditMessage() {
	var testCases = []struct {
		message   *chat.Message
		expectErr bool
	}{
		{message: &chat.Message{
			MessageID:   "em_message",
			SenderID:    "em_user",
			ChatID:      "em_chat",
			MessageText: "hellow world",
			TimeToPost:  tnow,
		}, expectErr: false},
		{message: &chat.Message{
			MessageID:   "em_message2",
			SenderID:    "em_user2",
			ChatID:      "em_chat",
			MessageText: "hellow world 2",
			TimeToPost:  time.Now(), // because of the time
		}, expectErr: true},
		{message: &chat.Message{
			MessageID:   "em_message",
			SenderID:    "em_user",
			ChatID:      "asjksdfhjksdhnjk",
			MessageText: "bye world",
			TimeToPost:  time.Now(),
		}, expectErr: true},
	}

	for _, v := range testCases {
		err := uts.repo.EditMessage(v.message)
		if v.expectErr {
			uts.Assert().Error(err, *v.message)
		} else {
			uts.Assert().Nil(err, *v.message)
		}
	}
}
