package tagrepository

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type tagRepository struct {
	db     *sql.DB
	gormDB *gorm.DB
}

func NewTagRepository(db *sql.DB, gormDB *gorm.DB) repository.TagRepositoryInterface {
	return &tagRepository{db, gormDB}
}

func (r *tagRepository) InsertTag(tag *entity.Tag) error {
	query := `INSERT INTO tags (serial, name) VALUES ($1, $2)`

	_, err := r.db.Exec(query, tag.Serial, tag.Name)
	if err != nil {
		return fmt.Errorf("error repo insert tag: %v", err.Error())
	}

	return nil
}

func (r *tagRepository) GetTags(pagination *entity.Pagination) ([]*entity.TagDetail, error) {
	tagDetails := []*entity.TagDetail{}

	db := r.gormDB.Table("tags t").
		Joins("LEFT JOIN tag_stats ts ON ts.tag_serial = t.serial")

	var total int64
	if pagination != nil {
		if err := db.Count(&total).Error; err != nil {
			return nil, fmt.Errorf("error repo get articles: %s", err.Error())
		}
		if total == 0 {
			return nil, nil
		}
		pagination.Total = int(total)

		pagination.SetPagination()
		limit := pagination.PageSize
		offset := pagination.GetOffset()

		db = db.Select("t.serial, t.name, ts.usage_count, ts.trending_score").Limit(int(limit)).Offset(int(offset)).Order("created_at DESC")
	}

	if err := db.Scan(&tagDetails).Error; err != nil {
		return nil, fmt.Errorf("error repo get tags: %s", err.Error())
	}

	return tagDetails, nil
}

func (r *tagRepository) GetTagBySerial(serial string) (*entity.TagDetail, error) {
	tagDetail := &entity.TagDetail{}

	err := r.gormDB.Table("tags t").
		Select("t.serial, t.name, ts.usage_count, ts.trending_score").
		Where("t.serial = ?", serial).
		Joins("LEFT JOIN tag_stats ts ON ts.tag_serial = t.serial").Scan(&tagDetail).Error
	if err != nil {
		return nil, fmt.Errorf("error repo get tag by serial: %s", err.Error())
	}

	return tagDetail, nil
}
