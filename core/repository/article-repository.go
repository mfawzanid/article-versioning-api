package repository

import (
	"article-versioning-api/core/entity"
	"database/sql"

	"gorm.io/gorm"
)

type ArticleRepositoryInterface interface {
	GetDb() *sql.DB

	InsertArticleTx(tx *sql.Tx, article *entity.Article) error
	InsertVersionTx(tx *sql.Tx, version *entity.Version) error
	InsertVersionTagsTx(tx *sql.Tx, versionSerial string, tagSerials []string) error
	UpdateArticleVersionStatus(tx *gorm.DB, req *entity.UpdateArticleVersionStatusRequest) error
	DeleteArticle(tx *gorm.DB, serial string) error
	DeleteVersionByArticleSerial(tx *gorm.DB, articleSerial string) error
	GetLatestVersionNumber(articleSerial string) (int, error)
	GetArticles(req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error)
	GetArticleLatestDetail(articleSerial string) ([]*entity.Version, error)
	GetVersionsByQuery(req *entity.GetVersionsByQueryRequest) ([]*entity.Version, error)
	GetVersionBySerial(serial string) (*entity.Version, error)
}
