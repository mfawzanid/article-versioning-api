package articlerepository

import (
	"article-versioning-api/config"
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	transactionutil "article-versioning-api/utils/transaction"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type articleRepository struct {
	db     *sql.DB
	cfg    *config.Config
	gormDB *gorm.DB
}

func NewArticleRepository(db *sql.DB, cfg *config.Config, gormDB *gorm.DB) repository.ArticleRepositoryInterface {
	return &articleRepository{db, cfg, gormDB}
}

func (r *articleRepository) GetDb() *sql.DB {
	return r.db
}

func (r *articleRepository) InsertArticleTx(tx *sql.Tx, article *entity.Article) error {
	query := `INSERT INTO articles (serial) VALUES ($1)`

	_, err := tx.Exec(query, article.Serial)
	if err != nil {
		return fmt.Errorf("error repo insert article: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) InsertVersionTx(tx *sql.Tx, version *entity.Version) error {
	query := `INSERT INTO versions (serial, author_username, version_number, article_serial, status, title, content) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(query, version.Serial, version.AuthorUsername, version.VersionNumber, version.ArticleSerial, version.Status, version.Title, version.Content)
	if err != nil {
		return fmt.Errorf("error repo insert version: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) InsertVersionTagsTx(tx *sql.Tx, versionSerial string, tagSerials []string) error {
	if len(tagSerials) == 0 {
		return nil
	}

	query := `INSERT INTO version_tags (version_serial, tag_serial) VALUES %s`

	values := []interface{}{}
	placeholders := []string{}

	for i, tagSerial := range tagSerials {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		values = append(values, versionSerial, tagSerial)
	}

	var queryValues string
	queryValues += strings.Join(placeholders, ", ")

	query = fmt.Sprintf(query, queryValues)

	_, err := tx.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("error repo insert version tag: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) UpdateArticleVersionStatus(tx *gorm.DB, req *entity.UpdateArticleVersionStatusRequest) error {
	conn := transactionutil.GetTransaction(tx)
	if conn == nil {
		conn = r.gormDB
	}

	// TODO: set published_at if status = published
	// and also remove for unpublishing
	query := `UPDATE versions SET status = ?, updated_at = NOW() WHERE serial = ? AND article_serial = ?`

	err := conn.Exec(query, req.NewStatus, req.VersionSerial, req.ArticleSerial).Error
	if err != nil {
		return fmt.Errorf("error repo update article version status: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) DeleteArticle(tx *gorm.DB, serial string) error {
	conn := transactionutil.GetTransaction(tx)
	if conn == nil {
		conn = r.gormDB
	}

	query := `UPDATE articles SET deleted_at = NOW(), updated_at = NOW() WHERE serial = ?`

	err := conn.Exec(query, serial).Error
	if err != nil {
		return fmt.Errorf("error repo delete article: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) DeleteVersionByArticleSerial(tx *gorm.DB, articleSerial string) error {
	conn := transactionutil.GetTransaction(tx)
	if conn == nil {
		conn = r.gormDB
	}

	query := `UPDATE versions SET status = ?, deleted_at = NOW(), updated_at = NOW(), published_at = NULL WHERE article_serial = ?`

	err := conn.Exec(query, entity.VersionStatusDeleted.String(), articleSerial).Error
	if err != nil {
		return fmt.Errorf("error repo delete version: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) GetLatestVersionNumber(articleSerial string) (int, error) {
	query := `SELECT MAX(v.version_number) AS latest_version_number 
				FROM articles a 
				INNER JOIN versions v ON a.serial = v.article_serial
				WHERE a.serial = $1
				GROUP BY a.serial`

	var latestVersionNumber *int

	err := r.db.QueryRow(query, articleSerial).Scan(&latestVersionNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error repo get latest version number: not found"))
		} else {
			return 0, fmt.Errorf("error repo get latest version number: %v", err.Error())
		}
	}

	return *latestVersionNumber, nil
}

func (r *articleRepository) GetArticles(req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error) {
	dtoVersions := []*Version{}

	db := r.gormDB.Table("versions v").
		Joins("INNER JOIN articles a ON a.serial = v.article_serial").
		Where("a.deleted_at IS NULL AND v.status = ?", req.Status)

	if req.AuthorUsername != "" {
		db = db.Where("v.author_usename = ?", req.AuthorUsername)
	}
	if req.TagSerial != "" {
		db = db.Joins("INNER JOIN version_tags vt ON vt.version_serial = v.serial").
			Where("vt.tag_serial = ?", req.TagSerial)
	}

	var total int64
	if req.Pagination != nil {
		if err := db.Count(&total).Error; err != nil {
			return nil, fmt.Errorf("error repo get articles: %s", err.Error())
		}
		if total == 0 {
			return &entity.GetArticlesResponse{
				Versions:   []*entity.Version{},
				Pagination: &entity.Pagination{},
			}, nil
		}
		req.Pagination.Total = int(total)

		req.Pagination.SetPagination()
		limit := req.Pagination.PageSize
		offset := req.Pagination.GetOffset()

		sortBy, sortType := sanitizeSort(req.SortBy, req.SortType)

		db = db.Limit(int(limit)).Offset(int(offset)).Order(fmt.Sprint(sortBy, " ", sortType)).Select("v.*")
	}

	if err := db.Scan(&dtoVersions).Error; err != nil {
		return nil, fmt.Errorf("error repo get articles: %s", err.Error())
	}

	versions, err := r.parseDTOToVersions(dtoVersions)
	if err != nil {
		return nil, fmt.Errorf("error repo get articles: %s", err.Error())
	}

	return &entity.GetArticlesResponse{
		Versions:   versions,
		Pagination: req.Pagination,
	}, nil
}

// get published version and version with latest version number
func (r *articleRepository) GetArticleLatestDetail(articleSerial string) ([]*entity.Version, error) {
	query := `
		(SELECT *, 1 AS sort_order -- make sure the first versions is the published one
		FROM versions
		WHERE article_serial = ? AND status = ?
		ORDER BY published_at DESC, version_number DESC
		LIMIT 1)

		UNION ALL

		(SELECT *, 2 AS sort_order -- make sure the second versions is the latest version number
		FROM versions
		WHERE article_serial = ? 
		ORDER BY version_number DESC
		LIMIT 1)
	`

	dtoVersions := []*Version{}
	if err := r.gormDB.Raw(query, articleSerial, entity.VersionStatusPublished.String(), articleSerial).Scan(&dtoVersions).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error repo get article latest detail: %s", err))
		}
		return nil, fmt.Errorf("error repo get article latest detail: %s", err)
	}

	versions, err := r.parseDTOToVersions(dtoVersions)
	if err != nil {
		return nil, fmt.Errorf("error repo get article latest detail: %s", err)
	}

	return versions, nil
}

func (r *articleRepository) GetVersionsByQuery(req *entity.GetVersionsByQueryRequest) ([]*entity.Version, error) {
	dtoVersions := []*Version{}

	db := r.gormDB.Table("versions")

	if req.ArticleSerial != "" {
		db = db.Where("article_serial = ?", req.ArticleSerial)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	err := db.Order("version_number ASC").Scan(&dtoVersions).Error
	if err != nil {
		return nil, fmt.Errorf("error repo get versions by query: %s", err.Error())
	}

	versions, err := r.parseDTOToVersions(dtoVersions)
	if err != nil {
		return nil, fmt.Errorf("error repo get versions by query: %s", err)
	}

	return versions, nil
}

func (r *articleRepository) GetVersionBySerial(serial string) (*entity.Version, error) {
	dtoVersions := []*Version{}

	err := r.gormDB.Table("versions").
		Where("serial = ?", serial).
		Scan(&dtoVersions).Error
	if err != nil {
		return nil, fmt.Errorf("error repo get version by serial: %s", err.Error())
	}

	versions, err := r.parseDTOToVersions(dtoVersions)
	if err != nil {
		return nil, fmt.Errorf("error repo get version by serial: %s", err)
	}
	if len(versions) == 0 {
		return nil, nil
	}

	return versions[0], nil
}

func sanitizeSort(sortBy, sortType string) (string, string) {
	var allowedSortBy = map[string]string{
		"created_at":             "a.created_at",
		"updated_at":             "a.updated_at",
		"published_at":           "a.published_at",
		"tag_relationship_score": "v.tag_relationship_score",
	}

	var allowedSortType = map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}

	sortBy, ok := allowedSortBy[sortBy]
	if !ok {
		sortBy = allowedSortBy["created_at"] // default
	}

	sortType, ok = allowedSortType[strings.ToLower(sortType)]
	if !ok {
		sortType = allowedSortType["desc"] // default
	}

	return sortBy, sortType
}

// parse dto to versions, include the tags
func (r *articleRepository) parseDTOToVersions(dtoVersions []*Version) ([]*entity.Version, error) {
	versions := []*entity.Version{}

	if len(dtoVersions) == 0 {
		return versions, nil
	}

	var versionSerials []string
	for _, v := range dtoVersions {
		versionSerials = append(versionSerials, v.Serial)

		versions = append(versions, v.parseToVersion())
	}

	var dtoVersionTags []*VersionTag
	err := r.gormDB.Table("version_tags vt").
		Select("vt.version_serial, vt.tag_serial, t.name AS tag_name").
		Joins("INNER JOIN tags t ON vt.tag_serial = t.serial").
		Where("vt.version_serial IN ?", versionSerials).
		Find(&dtoVersionTags).Error
	if err != nil {
		return nil, fmt.Errorf("error repo get version tag: %s", err.Error())
	}

	tagMap := make(map[string][]*VersionTag)
	for _, vt := range dtoVersionTags {
		tagMap[vt.VersionSerial] = append(tagMap[vt.VersionSerial], vt)
	}

	for _, version := range versions {
		if vts, ok := tagMap[version.Serial]; ok {
			tags := []*entity.Tag{}
			for _, vt := range vts {
				tags = append(tags, &entity.Tag{
					Serial: vt.TagSerial,
					Name:   vt.TagName,
				})
			}
			version.Tags = tags
		}
	}

	return versions, nil
}

func (r *articleRepository) UpdateTagRelationshipScore(tx *gorm.DB, versionSerial string, tagRelationshipScore float32) error {
	conn := transactionutil.GetTransaction(tx)
	if conn == nil {
		conn = r.gormDB
	}

	query := `UPDATE versions SET tag_relationship_score = ?, updated_at = NOW() WHERE serial = ?`

	err := conn.Exec(query, tagRelationshipScore, versionSerial).Error
	if err != nil {
		return fmt.Errorf("error repo update article version status: %v", err.Error())
	}

	return nil
}

func (r *articleRepository) GetTotalPublishedArticle(tx *gorm.DB) (int, error) {
	conn := transactionutil.GetTransaction(tx)
	if conn == nil {
		conn = r.gormDB
	}

	var total int64
	err := conn.Table("versions").
		Where("status = ?", entity.VersionStatusPublished.String()).
		Distinct("article_serial").
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("error repo get total pubslished article: %s", err.Error())
	}

	return int(total), nil
}
