package repository

import (
	"article-versioning-api/core/entity"

	"gorm.io/gorm"
)

type TagRepositoryInterface interface {
	InsertTag(tag *entity.Tag, tx *gorm.DB) error
	GetTags(pg *entity.Pagination) ([]*entity.TagDetail, error)
	GetTagBySerial(serial string) (*entity.TagDetail, error)

	InsertTagStat(tagSerial string, tx *gorm.DB) error
	DecrementUsageCount(tagSerials []string, tx *gorm.DB) error 
	IncrementUsageCount(tagSerials []string, tx *gorm.DB) error 
}
