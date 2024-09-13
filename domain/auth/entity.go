package auth

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserId       string
	FirstName    string
	LastName     string
	Username     string
	MobileNumber string
	Email        string
	HashPassword string
}

type UserOpt func(*User)

func NewUser(email string, password string, options ...UserOpt) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{
		UserId:       uuid.New().String(),
		Username:     email,
		HashPassword: string(hash),
	}

	user, err = user.EditUser(options...)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) EditUser(options ...UserOpt) (*User, error) {
	if len(options) == 0 {
		return u, errors.New("there is no options to operate")
	}
	for _, opt := range options {
		opt(u)
	}

	return u, nil
}

func WithUserName(userName string) UserOpt {
	return func(u *User) {
		u.Username = userName
	}
}

func WithFirstName(first string) UserOpt {
	return func(u *User) {
		u.FirstName = first
	}
}

func WithLastName(last string) UserOpt {
	return func(u *User) {
		u.LastName = last
	}
}

func WithPhoneNumber(phone string) UserOpt {
	return func(u *User) {
		u.MobileNumber = phone
	}
}
