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
	DecrementUsageCount(tx *gorm.DB, tagSerials []string) error
	IncrementUsageCount(tx *gorm.DB, tagSerials []string) error
	GetTagStatsBySerials(tx *gorm.DB, serials []string) ([]*entity.TagStat, error)
	UpdateTagStat(tx *gorm.DB, tagSerial string, trendingScore float32) error
	IncrementTagPairStat(tx *gorm.DB, tag1Serial, tag2Serial string) error
	DecrementTagPairStat(tx *gorm.DB, tag1Serial, tag2Serial string) error
	GetTagPairStatsBySerials(tx *gorm.DB, serials []string) ([]*entity.TagPairStat, error)
	GetTagStats(tx *gorm.DB, pg *entity.Pagination) ([]*entity.TagStat, error)
}
