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
	UpdateArticleVersionStatus(req *entity.UpdateArticleVersionStatusRequest, conn *gorm.DB) error
	DeleteArticle(tx *sql.Tx, serial string) error
	DeleteVersionByArticleSerial(tx *sql.Tx, articleSerial string) error
	GetLatestVersionNumber(articleSerial string) (int, error)
	GetArticles(req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error)
	GetArticleLatestDetail(articleSerial string) ([]*entity.Version, error)
	GetVersionsByQuery(req *entity.GetVersionsByQueryRequest) ([]*entity.Version, error)
	GetVersionBySerial(serial string) (*entity.Version, error)
}
