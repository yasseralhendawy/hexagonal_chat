package claims

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yasseralhendawy/hexagonal_chat/config"
)

type TokenGenerator struct {
	cfg *config.JWT
}

func NewTokenGenerator(cfg *config.JWT) *TokenGenerator {
	return &TokenGenerator{
		cfg: cfg,
	}
}

func (t *TokenGenerator) CreateToken(data *ClaimsData) (string, error) {
	claims := t.newClaims(data)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.cfg.Secret)
}

func (t *TokenGenerator) ValidateToken(tokenString string) (*ClaimsData, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.cfg.Secret), nil
	}, jwt.WithLeeway(t.cfg.LeewayInSecond*time.Second))
	if err != nil {
		return nil, err
	} else if claims, ok := token.Claims.(*claims); ok {
		return &claims.Date, nil
	} else {
		return nil, errors.New("unknown claims type, cannot proceed")
	}
}

// to be implement with redis
func (t *TokenGenerator) RefreshToken(tokenString string) (string, error) {
	return "", nil
}

func (t *TokenGenerator) newClaims(data *ClaimsData) *claims {
	return &claims{
		*data,
		jwt.RegisteredClaims{
			// ID:        data.UserID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.cfg.ExpireInHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    t.cfg.Issuer,
			Subject:   data.UserID,
			Audience:  t.cfg.Audience,
		},
	}
}
