package auth

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginMethod string

const (
	Email LoginMethod = "email"
)

type User struct {
	UserId       string
	FirstName    string
	LastName     string
	Username     string
	MobileNumber string
	Email        string
	HashPassword string
	LoginMethod  []LoginMethod
}

type UserOpt func(*User)

func NewUser(email string, password string, options ...UserOpt) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &User{
		UserId:       uuid.New().String(),
		Email:        email,
		LoginMethod:  []LoginMethod{Email},
		HashPassword: string(hash),
	}
	user = user.EditUser(options...)

	return user, nil
}

func (u *User) EditUser(options ...UserOpt) *User {
	if len(options) == 0 {
		return u
	}
	for _, opt := range options {
		opt(u)
	}

	return u
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

func WithLoginMethod(methods []LoginMethod) UserOpt {
	return func(u *User) {
		u.LoginMethod = methods
	}
}
