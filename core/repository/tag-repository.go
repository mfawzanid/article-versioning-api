package repository

import "article-versioning-api/core/entity"

type TagRepositoryInterface interface {
	InsertTag(tag *entity.Tag) error
	GetTags(pg *entity.Pagination) ([]*entity.TagDetail, error)
	GetTagBySerial(serial string) (*entity.TagDetail, error)
}
