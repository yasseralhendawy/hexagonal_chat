package cstorage_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	cstorage "github.com/yasseralhendawy/hexagonal_chat/internal/infrastructure/cassandra"
	appmetrics "github.com/yasseralhendawy/hexagonal_chat/pkg/metrics/adapter"
)

type _AuthRTestSuit struct {
	suite.Suite
	session *cstorage.CassandraDB
	repo    *cstorage.AuthRepo
}

func (uts *_AuthRTestSuit) SetupSuite() {
	var err error
	uts.session, err = cstorage.NewCassandraSession(cfg)
	uts.Require().Nil(err)
	mock_metric := &appmetrics.MockMetrics{}
	mock_metric.EXPECT().DBCallsWithLabelValues(mock.Anything, mock.Anything, mock.Anything).Return()
	uts.repo, err = uts.session.NewAuthRepo(mock_metric)
	uts.Require().Nil(err)
}

func (uts *_AuthRTestSuit) BeforeTest(suiteName, testName string) {
	fmt.Printf("Before running %s.%s \n", suiteName, testName)

	switch testName {
	case "TestCheckUserEmailExist": //lets add that user which suppose to be found
		err := uts.repo.CreateNewUser(&auth.User{UserId: "id1", Email: "check@mail.com", Username: "check", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
	case "TestGetUserByEmail":
		err := uts.repo.CreateNewUser(&auth.User{UserId: "id2", Email: "get@mail.com", Username: "get", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
	case "TestEditUser":
		err := uts.repo.CreateNewUser(&auth.User{UserId: "id3", Email: "edit@mail.com", Username: "edit", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}})
		uts.Require().Nil(err)
	default:
	}
}

func (uts *_AuthRTestSuit) TearDownSuite() {
	err := uts.session.DropTables()
	uts.Require().Nil(err)
}

func Test_AuthRTestSuit(t *testing.T) {
	suite.Run(t, &_AuthRTestSuit{})
}

func (uts *_AuthRTestSuit) TestCreateNewUser() {
	var testCases = []struct {
		user *auth.User
		err  bool
	}{
		{&auth.User{UserId: "279e0eb8-f175-4008-a9a6-b6e0852d9aee", Email: "user1@mail.com", Username: "user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32"}, true},
		{&auth.User{UserId: "279e0eb8-f175-4008-a9a6-b6e0852d9aee", Email: "user1@mail.com", Username: "user", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}}, false},
	}

	for _, tc := range testCases {
		err := uts.repo.CreateNewUser(tc.user)
		if tc.err {
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
		}
	}
}

func (uts *_AuthRTestSuit) TestCheckUserEmailExist() {
	var testCases = []struct {
		email string
		found bool
	}{
		{"user@mail.com", false},
		{"check@mail.com", true},
	}
	for _, tc := range testCases {
		found, err := uts.repo.CheckUserEmailExist(tc.email)
		uts.Assert().NoError(err)
		uts.Assert().Equal(tc.found, found)
	}
}

func (uts *_AuthRTestSuit) TestGetUserByEmail() {
	var testCases = []struct {
		email string
		user  *auth.User
		found bool
	}{
		{email: "lol@mail.com", user: nil, found: false},
		{email: "get@mail.com", user: &auth.User{UserId: "id2", Email: "get@mail.com", Username: "get", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}}, found: true},
	}

	for _, tc := range testCases {
		user, err := uts.repo.GetUserByEmail(tc.email)
		if tc.found {
			uts.Assert().Equal(tc.user, user)
			uts.Assert().Nil(err)
		} else {
			uts.Assert().Nil(user)
			uts.Assert().Error(err)
		}
	}
}

func (uts *_AuthRTestSuit) TestEditUser() {
	var testCases = []struct {
		user *auth.User
		err  bool
	}{
		{user: nil, err: true},
		{user: &auth.User{UserId: "lol", Email: "lol@mail.com", Username: "lol"}, err: true},
		{user: &auth.User{UserId: "id3", Email: "edit@mail.com", Username: "edit", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32", LoginMethod: []auth.LoginMethod{auth.Email}}, err: false},
	}

	for _, tc := range testCases {
		err := uts.repo.EditUser(tc.user)
		if tc.err {
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
		}
	}
}
