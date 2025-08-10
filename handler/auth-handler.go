package handler

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/usecase"
	errorutil "article-versioning-api/utils/error"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	VerifyToken(ctx *gin.Context)
	VerifyRole(authorizedRoles []string) gin.HandlerFunc
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
			errorutil.Error: http.StatusText(http.StatusUnauthorized),
		})
		ctx.Abort()
		return
	}

	user, err := h.authUsecase.VerifyToken(authToken[0])
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			errorutil.Error: http.StatusText(http.StatusUnauthorized),
		})
		ctx.Abort()
		return
	}

	ctx.Set(entity.ContextUsername, user.Username)
	ctx.Set(entity.ContextRole, user.Role)

	ctx.Next()
}

func (h *authHandler) VerifyRole(authorizedRoles []string) gin.HandlerFunc {
	return func(ctx *gin.Context){
		userRole, _ := ctx.Get(entity.ContextRole)
		for _, role := range authorizedRoles {
			if role == userRole {
				ctx.Next()
				return
			}
		}

		ctx.JSON(http.StatusUnauthorized, gin.H{
			message: errorutil.CombineHTTPErrorMessage(http.StatusUnauthorized, errors.New("role is not allowed")),
		})
		ctx.Abort()

	}
}
