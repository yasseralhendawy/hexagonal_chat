package cstorage_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	"github.com/yasseralhendawy/hexagonal_chat/domain/user"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type _UserChatRepoSuite struct {
	suite.Suite
	session  *cstorage.CassandraDB
	repo     *cstorage.UserChatRepo
	authRepo *cstorage.AuthRepo
}

func (uts *_UserChatRepoSuite) SetupSuite() {
	var err error
	uts.session, err = cstorage.NewCassandraSession(cfg)
	uts.Require().Nil(err)
	metric := appmetrics.MockMetrics{}
	metric.EXPECT().DBCallsWithLabelValues(mock.Anything, mock.Anything, mock.Anything).Return()
	uts.repo, err = uts.session.NewUserChatRepo(&metric)
	uts.Require().Nil(err)
	uts.authRepo, err = uts.session.NewAuthRepo(&metric)
	uts.Require().Nil(err)
}

func (uts *_UserChatRepoSuite) BeforeTest(suiteName, testName string) {
	fmt.Printf("Before running %s.%s \n", suiteName, testName)
	switch testName {
	case "TestCheckParticipants":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "cp_user", Email: "check@mail.com", Username: "cp_user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
	case "TestCreateNewChat":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "user1", Email: "check@mail.com", Username: "user1", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.authRepo.CreateNewUser(&auth.User{UserId: "user2", Email: "user2@mail.com", Username: "user2", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
	case "TestGetUserChat":
		err := uts.repo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "user1", Username: "user1"},
		},
			ChatID: "g_chat",
		})
		uts.Require().Nil(err)
	case "TestGetUserHistory":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "user3", Email: "user3@mail.com", Username: "user3", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.repo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "user3", Username: "user3"},
		},
			ChatID: "guh_chat",
		})
		uts.Require().Nil(err)
	case "TestSaveUserChat":
		err := uts.authRepo.CreateNewUser(&auth.User{UserId: "user4", Email: "user4@mail.com", Username: "user4", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.authRepo.CreateNewUser(&auth.User{UserId: "user5", Email: "user4@mail.com", Username: "user5", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
		err = uts.repo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "user4", Username: "user4"},
		},
			ChatID: "suc_chat1",
		})
		uts.Require().Nil(err)
		err = uts.repo.CreateNewChat(&user.UserChat{Participants: []*user.Person{
			{PersonId: "user4", Username: "user4"},
			{PersonId: "user5", Username: "user5"},
		},
			ChatID: "suc_chat2",
		})
		uts.Require().Nil(err)
	default:
	}

}

func (uts *_UserChatRepoSuite) TearDownSuite() {
	err := uts.session.DropTables()
	uts.Require().Nil(err)
}
func Test_UserChatRepoSuite(t *testing.T) {
	suite.Run(t, &_UserChatRepoSuite{})
}

func (uts *_UserChatRepoSuite) TestCheckParticipants() {
	var testCases = []struct {
		listOfIds []string
		expectErr bool
	}{
		{[]string{"cp_user"}, false},         // user exist
		{[]string{}, true},                   // empty list
		{[]string{"cp_user", "asdas"}, true}, //mix of users dont and do exost
		{[]string{"asdsadf"}, true},          //user don't exist
	}
	for _, v := range testCases {
		_, err := uts.repo.CheckParticipants(v.listOfIds)
		if v.expectErr {
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
		}
	}

}

func (uts *_UserChatRepoSuite) TestCreateNewChat() {
	var testCases = []struct {
		testId    int16
		userChat  user.UserChat
		expectErr bool
	}{
		//with chatID
		{testId: 1,
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user1", Username: "user1"},
			},
				ChatID: "chat1",
			}, expectErr: false},
		{testId: 2,
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user2", Username: "user2"},
			},
				ChatID: "chat2",
			}, expectErr: false},
		//without chat id
		{testId: 3,
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user1", Username: "user1"},
			},
			}, expectErr: true},
	}

	for _, v := range testCases {
		err := uts.repo.CreateNewChat(&v.userChat)
		if v.expectErr {
			uts.Assert().Error(err)
			fmt.Println(err)
		} else {
			uts.Assert().Nil(err, "testId: ", v.testId)
		}
	}
	//with chatID
	//without chat id
}

func (uts *_UserChatRepoSuite) TestGetUserChat() {
	var testCases = []struct {
		chatID    string
		expectErr bool
	}{
		{chatID: "g_chat", expectErr: false},  //get one already exist
		{chatID: "asdgjkdb", expectErr: true}, //get one which is not exist
	}
	for _, v := range testCases {
		res, err := uts.repo.GetUserChat(v.chatID)
		if v.expectErr {
			uts.Assert().Error(err)
			uts.Assert().Nil(res)
		} else {
			uts.Assert().Nil(err)
			uts.Assert().NotNil(res)
			fmt.Println(res)
		}
	}
}

func (uts *_UserChatRepoSuite) TestGetUserHistory() {
	var testCases = []struct {
		userId    string
		len       int
		expectErr bool
	}{
		{userId: "user3", expectErr: false, len: 1},
		{userId: "sdhbgsdjbjjkbk", expectErr: false, len: 0},
	}
	for _, v := range testCases {
		res, err := uts.repo.GetUserHistory(v.userId)
		if v.expectErr {
			uts.Assert().Error(err)
			uts.Nil(res)
		} else {
			uts.Assert().Nil(err, v.userId)
			uts.Assert().Equal(len(res), v.len)
		}
	}
}

func (uts *_UserChatRepoSuite) TestSaveUserChat() {
	var testcases = []struct {
		userChat  user.UserChat
		opt       []user.ChatOpt
		expectErr bool
	}{
		{
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user1", Username: "user1"},
			},
				ChatID: "suc_chat_not_exist",
			}, opt: []user.ChatOpt{
				user.AddParticipants([]*user.Person{
					{PersonId: "user2", Username: "user2"},
				}),
			}, expectErr: true, // as there is no chat with chatID= "suc_chat_not_exist"
		},

		{
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user4", Username: "user4"},
			},
				ChatID: "suc_chat1",
			}, opt: []user.ChatOpt{
				user.AddParticipants([]*user.Person{
					{PersonId: "user5", Username: "user5"},
				}),
			}, expectErr: false, // test AddParticipants
		},
		{
			userChat: user.UserChat{Participants: []*user.Person{
				{PersonId: "user4", Username: "user4"},
				{PersonId: "user5", Username: "user5"},
			},
				ChatID: "suc_chat2",
			}, opt: []user.ChatOpt{
				user.RemoveParticipant("user5"),
			}, expectErr: false, // test RemoveParticipant
		},
	}

	for _, v := range testcases {
		c, err := v.userChat.EditChat(v.opt...)
		uts.Require().Nil(err)
		err = uts.repo.SaveUserChat(c)
		if v.expectErr {
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
		}
	}
}
