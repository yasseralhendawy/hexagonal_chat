package claims

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
)

type ClaimsData struct {
	UserID string `json:"userID"`
	Email  string `json:"email"`
}

func NewClaimsData(user *auth.User) *ClaimsData {
	return &ClaimsData{
		UserID: user.UserId,
		Email:  user.Email,
	}
}

type claims struct {
	Date ClaimsData `json:"claimsData"`
	jwt.RegisteredClaims
}
