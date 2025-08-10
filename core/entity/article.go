package entity

import (
	errorutil "article-versioning-api/utils/error"
	"time"

	"github.com/pkg/errors"
)

const (
	SortByCreatedAt            = "created_at"
	SortByUpdatedAt            = "updated_at"
	SortByPublishedAt          = "published_at"
	SortByTagRelationshipScore = "tag_relationship_score"

	SortTypeAsc  = "asc"
	SortTypeDesc = "desc"
)

type CreateArticleRequest struct {
	Title      string
	Content    string
	TagSerials []string
}

func (r *CreateArticleRequest) Validate() error {
	if r.Title == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article request: title is mandatory"))
	}
	if r.Content == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article request: content is mandatory"))
	}

	return nil
}

type VersionStatus int

const (
	VersionStatusUnknown VersionStatus = iota
	VersionStatusDraft
	VersionStatusPublished
	VersionStatusArchived
	VersionStatusDeleted
)

var (
	mapVersionStatusToString = map[VersionStatus]string{
		VersionStatusUnknown:   "unknown",
		VersionStatusDraft:     "draft",
		VersionStatusPublished: "published",
		VersionStatusArchived:  "archived",
		VersionStatusDeleted:   "deleted",
	}
	mapStringToVersionStatus = map[string]VersionStatus{
		"unknown":   VersionStatusUnknown,
		"draft":     VersionStatusDraft,
		"published": VersionStatusPublished,
		"archived":  VersionStatusArchived,
		"deleted":   VersionStatusDeleted,
	}
)

func (vs VersionStatus) String() string {
	return mapVersionStatusToString[vs]
}

func StringToVersionRole(vs string) VersionStatus {
	versionStatus, ok := mapStringToVersionStatus[vs]
	if !ok {
		return VersionStatusUnknown
	}

	return versionStatus
}

type Version struct {
	Serial               string     `json:"serial"`
	AuthorUsername       string     `json:"authorUsername"`
	VersionNumber        int        `json:"versionNumber"`
	ArticleSerial        string     `json:"articleSerial"`
	Title                string     `json:"title"`
	Content              string     `json:"content"`
	Status               string     `json:"status"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            *time.Time `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt"`
	PublishedAt          *time.Time `json:"publishedAt"`
	TagRelationshipScore float32    `json:"tagRelationshipScore"`
	Tags                 []*Tag     `json:"tags"`
}

type CreateArticleResponse struct {
	ArticleSerial  string   `json:"articleSerial"`
	AuthorUsername string   `json:"authorUsername"`
	Version        *Version `json:"version"`
}

type Article struct {
	Serial   string
	Versions []*Version
}

type VersionTag struct {
	VersionSerial string
	TagSerial     string
}

type UpdateArticleVersionStatusRequest struct {
	ArticleSerial string
	VersionSerial string
	NewStatus     string
}

func (r *UpdateArticleVersionStatusRequest) Validate() error {
	if r.ArticleSerial == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error update article version status: article serial is mandatory"))
	}
	if r.VersionSerial == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error update article version status: version serial is mandatory"))
	}
	if StringToVersionRole(r.NewStatus) == VersionStatusUnknown {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error update article version status: version status is unknown"))
	}

	return nil
}

type CreateArticleVersionRequest struct {
	ArticleSerial string
	Title         string
	Content       string
	TagSerials    []string
}

func (r *CreateArticleVersionRequest) Validate() error {
	if r.ArticleSerial == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article version request: article serial is mandatory"))
	}
	if r.Title == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article version request: title is mandatory"))
	}
	if r.Content == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article version request: content is mandatory"))
	}

	return nil
}

type CreateArticleVersionResponse struct {
	ArticleSerial string   `json:"articleSerial"`
	AuthorId      string   `json:"authorId"`
	Version       *Version `json:"version"`
}

type GetArticlesRequest struct {
	Status         string `form:"status"`
	AuthorUsername string `form:"authorUsername"`
	TagSerial      string `form:"tagSerial"`
	Page           int    `form:"page"`
	PageSize       int    `form:"pageSize"`
	Pagination     *Pagination
	SortBy         string `form:"sortBy"`   // created_at, updated_at, published_at, tag_relationship_score
	SortType       string `form:"sortType"` // asc, desc
}

var (
	validSortBy = map[string]bool{
		SortByCreatedAt:            true,
		SortByUpdatedAt:            true,
		SortByPublishedAt:          true,
		SortByTagRelationshipScore: true,
	}
	validSortType = map[string]bool{
		SortTypeAsc:  true,
		SortTypeDesc: true,
	}
)

func (r *GetArticlesRequest) Validate() error {
	if r.Pagination != nil {
		r.Pagination.Validate()
	}
	if r.SortBy != "" && !validSortBy[r.SortBy] {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get artiles: sort by value is unknown"))
	}
	if r.SortType != "" && !validSortType[r.SortType] {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get artiles: sort type value is unknown"))
	}
	return nil
}

type GetArticlesResponse struct {
	Versions   []*Version  `json:"version"`
	Pagination *Pagination `json:"pagination"`
}

type GetArticleLatestDetailResponse struct {
	PublishedVersion *Version `json:"publishedVersion"`
	LatestVersion    *Version `json:"latestVersion"`
}

type GetVersionsByArticleSerialResponse struct {
	Versions []*Version `json:"versions"`
}

type GetTagsResponse struct {
	Tags       []*TagDetail `json:"tags"`
	Pagination *Pagination  `json:"pagination"`
}
