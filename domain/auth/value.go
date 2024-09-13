package auth

import "golang.org/x/crypto/bcrypt"

func (u *User) checkPasswordMatch(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.HashPassword), []byte(password))
}
