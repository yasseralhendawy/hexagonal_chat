package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
)

type MockTestSuit struct {
	suite.Suite

	service *auth.Service
	repo    *auth.MockIAuthRepo
}

func TestMockTestSuit(t *testing.T) {
	suite.Run(t, &MockTestSuit{})
}

func (mut *MockTestSuit) SetupTest() {
	mr := auth.MockIAuthRepo{}
	s := auth.New(&mr)
	mut.repo = &mr
	mut.service = s
}

func (uts *MockTestSuit) TestCreateNewUser() {
	var testCases = []struct {
		email            string
		password         string
		emailExist       bool
		emailExistErr    error
		mocCreateUserErr error
		expectErr        bool
	}{
		{"user@mail.com", "123456", false, nil, nil, false},
		{"user@mail.com", "123456", false, nil, errors.New("omg"), true},
		{"user@mail.com", "123456", true, errors.New("whatever error"), nil, true},
		{"user@mail.com", "123456", true, nil, nil, true},
	}

	for _, c := range testCases {
		uts.repo.EXPECT().CheckUserEmailExist(c.email).Return(c.emailExist, c.emailExistErr).Times(1)
		if !c.emailExist {
			uts.repo.EXPECT().CreateNewUser(mock.Anything).Return(c.mocCreateUserErr).Times(1)
		}
		res, err := uts.service.CreateNewUser(c.email, c.password)
		if c.expectErr {
			uts.Assert().Nil(res)
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
			uts.Assert().Equal(c.email, res.Email)
		}

	}
}

func (uts *MockTestSuit) TestGetUser_PasswordMatch() {
	var testCases = []struct {
		user      *auth.User
		email     string
		password  string
		mocErr    error
		expectErr bool
	}{
		{nil, "user@mail.com", "123456", errors.New("ok"), true},
		{&auth.User{Email: "user@mail.com", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32"},
			"user@mail.com", "123456", nil, false},
		{&auth.User{Email: "user@mail.com", HashPassword: "$2a$10$mDWm1y4VDrPm6nK8PKeIgOalw1wMcy.q8XbhgmbeL0eMm4r45jb32"},
			"user@mail.com", "1234566", nil, true},
	}

	for _, c := range testCases {
		uts.repo.EXPECT().GetUserByEmail(c.email).Return(c.user, c.mocErr).Times(1)
		res, err := uts.service.GetUser(c.email, c.password)
		if c.expectErr {
			uts.Assert().Nil(res)
			uts.Assert().Error(err)
		} else {
			uts.Assert().Nil(err)
			uts.Assert().Equal(c.email, res.Email)
		}

	}
}
