package entity

import (
	errorutil "article-versioning-api/utils/error"
	"time"

	"github.com/pkg/errors"
)

type CreateTagRequest struct {
	Name string
}

func (r *CreateTagRequest) Validate() error {
	if r.Name == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create tag request: name is mandatory"))
	}
	return nil
}

type GetTagsRequest struct {
	Page       int `form:"page"`
	PageSize   int `form:"pageSize"`
	Pagination *Pagination
}

type Tag struct {
	Serial string `json:"serial"`
	Name   string `json:"name"`
}

type TagDetail struct {
	Serial        string  `json:"serial"`
	Name          string  `json:"name"`
	UsageCount    int     `json:"usageCount"`
	TrendingScore float32 `json:"trendingScore"`
}

type TagStat struct {
	TagSerial              string
	UsageCount             float32
	TrendingScore          float32
	UsageCountUpdatedAt    *time.Time
	TrendingScoreUpdatedAt *time.Time
}

type TagPairStat struct {
	Tag1Serial string
	Tag2Serial string
	UsageCount int
}

type GetTagStatsRequest struct {
}
