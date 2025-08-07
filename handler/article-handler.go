package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type articleHandler struct {
}

func NewArticleHandler() *articleHandler {
	return &articleHandler{}
}

func (h *articleHandler) CreateArticle(c *gin.Context) {
	// TODO
	c.JSON(http.StatusOK, nil)
}
