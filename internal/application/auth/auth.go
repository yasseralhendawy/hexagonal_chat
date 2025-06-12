package authapp

import (
	"github.com/yasseralhendawy/hexagonal_chat/domain/auth"
	"github.com/yasseralhendawy/hexagonal_chat/pkg/claims"
)

type App struct {
	DomainService *auth.Service
	Tokenization  *claims.TokenGenerator
}

func (app *App) Login(req *LoginRequest) (any, error) {
	u, err := app.DomainService.GetUser(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := app.Tokenization.CreateToken(claims.NewClaimsData(u))
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (app *App) Register(req *RegisterRequest) (any, error) {
	u, err := app.DomainService.CreateNewUser(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := app.Tokenization.CreateToken(claims.NewClaimsData(u))
	if err != nil {
		return nil, err
	}
	return token, nil
}
