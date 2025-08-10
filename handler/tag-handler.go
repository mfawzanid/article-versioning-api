package handler

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/usecase"
	errorutil "article-versioning-api/utils/error"
	generalutil "article-versioning-api/utils/general"
	"net/http"

	"github.com/gin-gonic/gin"
)

type tagHandler struct {
	tagUsecase usecase.TagUsecaseInterface
}

func NewTagHandler(tagUsecase usecase.TagUsecaseInterface) *tagHandler {
	return &tagHandler{tagUsecase}
}

func (h *tagHandler) CreateTag(c *gin.Context) {
	req := &entity.CreateTagRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	serial, err := h.tagUsecase.CreateTag(req)
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
		"serial": serial,
	})
}

func (h *tagHandler) GetTags(c *gin.Context) {
	req := &entity.GetTagsRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			errorutil.Error: errorutil.CombineHTTPErrorMessage(http.StatusInternalServerError, err),
		})
		return
	}

	req.Pagination = entity.ParseToPagination(req.Page, req.PageSize) // TODO: move to usecase?

	resp, err := h.tagUsecase.GetTags(req)
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

func (h *tagHandler) GetTagBySerial(c *gin.Context) {
	serial, _ := c.Params.Get("serial")

	resp, err := h.tagUsecase.GetTagBySerial(serial)
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
