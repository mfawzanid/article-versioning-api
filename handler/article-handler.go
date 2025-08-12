package handler

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/usecase"
	errorutil "article-versioning-api/utils/error"
	generalutil "article-versioning-api/utils/general"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type articleHandler struct {
	articleUsecase usecase.ArticleUsecaseInterface
}

func NewArticleHandler(articleUsecase usecase.ArticleUsecaseInterface) *articleHandler {
	return &articleHandler{articleUsecase}
}

const (
	message = "message"
)

func (h *articleHandler) CreateArticle(c *gin.Context) {
	req := &entity.CreateArticleRequest{}

	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	resp, err := h.articleUsecase.CreateArticle(c, req)
	if err != nil {
		switch errorutil.GetErrorType(err) {
		case errorutil.ErrBadRequest:
			c.JSON(http.StatusBadRequest, generalutil.MapAny{
				errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusBadRequest, errorutil.GetOriginalError(err)),
			})
			return
		default:
			if c != nil { // TODO
				c.JSON(http.StatusInternalServerError, generalutil.MapAny{
					errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
				})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *articleHandler) GetArticles(c *gin.Context) {
	req := &entity.GetArticlesRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	req.Pagination = entity.ParseToPagination(req.Page, req.PageSize) // TODO: move to usecase?

	resp, err := h.articleUsecase.GetArticles(c, req)
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

	c.JSON(http.StatusOK, resp)
}

func (h *articleHandler) GetArticleLatestDetail(c *gin.Context) {
	articleSerial, _ := c.Params.Get("serial")

	resp, err := h.articleUsecase.GetArticleLatestDetail(articleSerial)
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

	c.JSON(http.StatusOK, resp)
}

func (h *articleHandler) CreateArticleVersion(c *gin.Context) {
	articleSerial, _ := c.Params.Get("serial")
	req := &entity.CreateArticleVersionRequest{}
	req.ArticleSerial = articleSerial

	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	resp, err := h.articleUsecase.CreateArticleVersion(c, req)
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

	c.JSON(http.StatusCreated, resp)
}

func (h *articleHandler) DeleteArticle(c *gin.Context) {
	articleSerial, _ := c.Params.Get("serial")

	err := h.articleUsecase.DeleteArticle(articleSerial)
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
		message: fmt.Sprintf("success delete article '%s'", articleSerial),
	})
}

func (h *articleHandler) UpdateArticleVersionStatus(c *gin.Context) {
	articleSerial, _ := c.Params.Get("serial")
	versionSerial, _ := c.Params.Get("versionSerial")

	req := &entity.UpdateArticleVersionStatusRequest{}
	req.ArticleSerial = articleSerial
	req.VersionSerial = versionSerial

	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	err := h.articleUsecase.UpdateArticleVersionStatus(req)
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
		message: fmt.Sprintf("success update status for version '%s'", versionSerial),
	})
}

func (h *articleHandler) GetVersionsByArticleSerial(c *gin.Context) {
	articleSerial, _ := c.Params.Get("serial")

	resp, err := h.articleUsecase.GetVersionsByArticleSerial(articleSerial)
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

	c.JSON(http.StatusOK, resp)
}

func (h *articleHandler) GetVersionBySerial(c *gin.Context) {
	versionSerial, _ := c.Params.Get("versionSerial")

	resp, err := h.articleUsecase.GetVersionBySerial(versionSerial)
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

	c.JSON(http.StatusOK, resp)
}

func (h *articleHandler) UpdateTrendingScoreTags(c *gin.Context) {
	pg := entity.Pagination{}
	err := h.articleUsecase.UpdateTrendingScoreTags(&pg)
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
		message: "tag trending score is updated",
	})
}
