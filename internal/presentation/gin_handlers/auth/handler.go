package ginhs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authapp "github.com/yasseralhendawy/hexagonal_chat/internal/application/auth"
	ginserver "github.com/yasseralhendawy/hexagonal_chat/pkg/servers/gin_server"
)

type AuthHandler struct {
	Server *ginserver.GinServer
	App    IAuthApp
}

type IAuthApp interface {
	Login(*authapp.LoginRequest) (interface{}, error)
	Register(*authapp.RegisterRequest) (interface{}, error)
}

func (h AuthHandler) Run(addr ...string) {
	h.Login()
	h.Register()

	h.Server.Run(addr...)
}

func (h AuthHandler) Login() {
	h.Server.Engin.GET("login", func(ctx *gin.Context) {
		req := new(authapp.LoginRequest)

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.App.Login(req)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, res)
	})
}

func (h AuthHandler) Register() {
	h.Server.Engin.GET("register", func(ctx *gin.Context) {
		req := new(authapp.RegisterRequest)

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.App.Register(req)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, res)
	})
}
