package handler

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	VerifyToken(ctx *gin.Context)
}

type authHandler struct {
	authUsecase usecase.AuthUsecaseInterface
}

func NewAuthHandler(authUsecase usecase.AuthUsecaseInterface) AuthHandler {
	return &authHandler{authUsecase}
}

func (h *authHandler) VerifyToken(ctx *gin.Context) {
	authToken := ctx.Request.Header["Authorization"]
	if len(authToken) == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": http.StatusText(http.StatusUnauthorized),
		})
		ctx.Abort()
		return
	}

	user, err := h.authUsecase.VerifyToken(authToken[0])
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": http.StatusText(http.StatusUnauthorized),
		})
		ctx.Abort()
		return
	}

	ctx.Set(entity.ContextUsername, user.Username)
	ctx.Set(entity.ContextRole, user.Role)

	ctx.Next()
}
