package handler

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/usecase"
	"net/http"

	errorutil "article-versioning-api/utils/error"
	generalutil "article-versioning-api/utils/general"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userUsecase usecase.UserUsecaseInterface
}

func NewUserHandler(userUsecase usecase.UserUsecaseInterface) *userHandler {
	return &userHandler{userUsecase}
}

func (h *userHandler) RegisterUser(c *gin.Context) {
	req := &entity.RegisterUserRequest{}

	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Message: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	err := h.userUsecase.RegisterUser(req)
	if err != nil {
		switch errorutil.GetErrorType(err) {
		case errorutil.ErrBadRequest:
			c.JSON(http.StatusBadRequest, generalutil.MapAny{
				errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusBadRequest, errorutil.GetOriginalError(err)),
			})
			return
		default:
			if c != nil {
				c.JSON(http.StatusInternalServerError, generalutil.MapAny{
					errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
				})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		errorutil.Message: "username created successfully",
	})
}

func (h *userHandler) Login(c *gin.Context) {
	req := &entity.LoginRequest{}

	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Message: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	token, err := h.userUsecase.Login(req)
	if err != nil {
		switch errorutil.GetErrorType(err) {
		case errorutil.ErrBadRequest:
			c.JSON(http.StatusBadRequest, generalutil.MapAny{
				errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusBadRequest, errorutil.GetOriginalError(err)),
			})
			return
		default:
			if c != nil {
				c.JSON(http.StatusInternalServerError, generalutil.MapAny{
					errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
				})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
