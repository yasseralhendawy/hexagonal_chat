package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
)

func TestNewUser(t *testing.T) {
	var testCases = []struct {
		email     string
		pass      string
		expectErr bool
	}{
		{"user@mail.com", "123456", false},
		{"user@email.com", "11111111111111111111111111111111111111111111111111111111111111111111111111111111", true},
	}
	assert := assert.New(t)
	for _, c := range testCases {
		user, err := auth.NewUser(c.email, c.pass)
		if c.expectErr {
			assert.Error(err)
			assert.Nil(user)
		} else {
			assert.Nil(err)
			assert.NotEqual(c.pass, user.HashPassword)
			assert.Equal(c.email, user.Email)
			assert.NotEmpty(user.HashPassword)
			assert.NotEmpty(user.UserId)
		}
	}
}

func TestEditUser(t *testing.T) {
	email := "user@mail.com"
	pass := "123456"
	username := "username"
	firstname := "firstname"
	lastname := "lastname"
	phone := "+201111111111"
	assert := assert.New(t)

	user, err := auth.NewUser(email, pass)
	assert.NoError(err)

	user = user.EditUser(auth.WithFirstName(firstname), auth.WithLastName(lastname), auth.WithUserName(username), auth.WithPhoneNumber(phone), auth.WithLoginMethod([]auth.LoginMethod{auth.Email}))

	assert.Equal(username, user.Username)
	assert.Equal(firstname, user.FirstName)
	assert.Equal(lastname, user.LastName)
	assert.Equal(phone, user.MobileNumber)
}
