package ginserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authapp "github.com/yasseralhendawy/hexagonal_chat/internal/application/auth"
)

type AuthHandler struct {
	Server *GinServer
	App    *authapp.App
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
	h.Server.Engin.POST("register", func(ctx *gin.Context) {
		req := new(authapp.RegisterRequest)

		err := ctx.ShouldBindJSON(&req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		res, err := h.App.Register(req)
		if err != nil {
			ctx.JSON(http.StatusConflict, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, res)
	})
}
