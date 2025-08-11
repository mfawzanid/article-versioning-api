package usecase

import (
	"article-versioning-api/config"
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	generalutil "article-versioning-api/utils/general"
	serialutil "article-versioning-api/utils/serial"
	transactionutil "article-versioning-api/utils/transaction"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type articleUsecase struct {
	articleRepo    repository.ArticleRepositoryInterface
	tagRepo        repository.TagRepositoryInterface
	transactionPkg transactionutil.Transaction
	cfg            *config.Config
}

type ArticleUsecaseInterface interface {
	CreateArticle(ctx *gin.Context, req *entity.CreateArticleRequest) (*entity.CreateArticleResponse, error)
	UpdateArticleVersionStatus(req *entity.UpdateArticleVersionStatusRequest) error
	DeleteArticle(articleSerial string) error
	CreateArticleVersion(ctx *gin.Context, req *entity.CreateArticleVersionRequest) (resp *entity.CreateArticleVersionResponse, err error)
	GetArticles(ctx *gin.Context, req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error)
	GetArticleLatestDetail(articleSerial string) (*entity.GetArticleLatestDetailResponse, error)
	GetVersionsByArticleSerial(articleSerial string) (*entity.GetVersionsByArticleSerialResponse, error)
	GetVersionBySerial(serial string) (*entity.Version, error)
}

func NewArticleUsecase(articleRepo repository.ArticleRepositoryInterface, tagRepo repository.TagRepositoryInterface, transactionPkg transactionutil.Transaction, cfg *config.Config) ArticleUsecaseInterface {
	return &articleUsecase{articleRepo, tagRepo, transactionPkg, cfg}
}

const (
	articleSerialPrefix = "ART"
	versionSerialPrefix = "VER"
)

func (u *articleUsecase) CreateArticle(ctx *gin.Context, req *entity.CreateArticleRequest) (resp *entity.CreateArticleResponse, err error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	authorUsername := entity.GetContextUsername(ctx)
	if authorUsername == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article: user id not found in context"))
	}

	tx, err := u.articleRepo.GetDb().Begin()
	if err != nil {
		return nil, fmt.Errorf("error create article: failed to begin transaction: %s", err.Error())
	}

	defer func() {
		err = transactionutil.SettleTransaction(tx, err)
	}()

	articleSerial, err := serialutil.GenerateId(articleSerialPrefix)
	if err != nil {
		return nil, fmt.Errorf("error create article: error generate serial: %s", err.Error())
	}

	article := &entity.Article{
		Serial: articleSerial,
	}

	err = u.articleRepo.InsertArticleTx(tx, article)
	if err != nil {
		return
	}

	versionSerial, err := serialutil.GenerateId(versionSerialPrefix)
	if err != nil {
		return nil, fmt.Errorf("error create article: error generate version serial: %s", err.Error())
	}
	version := &entity.Version{
		Serial:         versionSerial,
		AuthorUsername: authorUsername,
		VersionNumber:  1,
		ArticleSerial:  articleSerial,
		Title:          req.Title,
		Content:        req.Content,
		Status:         entity.VersionStatusDraft.String(),
	}

	err = u.articleRepo.InsertVersionTx(tx, version)
	if err != nil {
		return
	}

	err = u.articleRepo.InsertVersionTagsTx(tx, versionSerial, req.TagSerials)
	if err != nil {
		return
	}

	resp = &entity.CreateArticleResponse{
		ArticleSerial:  articleSerial,
		AuthorUsername: authorUsername,
		Version:        version,
	}

	return
}

func (u *articleUsecase) UpdateArticleVersionStatus(req *entity.UpdateArticleVersionStatusRequest) (err error) {
	err = req.Validate()
	if err != nil {
		return err
	}

	var currPublishedVersion *entity.Version
	if entity.IsPublishedStatus(req.NewStatus) {
		publishedVersions, err := u.articleRepo.GetVersionsByQuery(&entity.GetVersionsByQueryRequest{
			ArticleSerial: req.ArticleSerial,
			Status:        entity.VersionStatusPublished.String(),
		})
		if err != nil {
			return fmt.Errorf("error get published version: %s", err.Error())
		}
		if len(publishedVersions) > 0 {
			currPublishedVersion = publishedVersions[0]
		}
	}

	// init transaction for updating status and tag's usage count
	tx := u.transactionPkg.InitTransaction()
	defer func() {
		u.transactionPkg.SettleTransaction(tx, err)
	}()

	// calculate tag usage count
	version, err := u.articleRepo.GetVersionBySerial(req.VersionSerial)
	if err != nil {
		return err
	}

	currStatus := version.Status
	newStatus := req.NewStatus

	allAffectedTagSerials := []string{}

	if currStatus == newStatus {
		return nil
	} else if entity.IsPublishedStatus(currStatus) == entity.IsPublishedStatus(newStatus) || !entity.IsPublishedStatus(currStatus) == !entity.IsPublishedStatus(newStatus) {
		// published to published or non published to non published, then do nothing
		return nil
	} else {
		tagsSerials := version.TagSerials()

		allAffectedTagSerials = append([]string{}, tagsSerials...)

		if !entity.IsPublishedStatus(currStatus) && entity.IsPublishedStatus(newStatus) { // publish
			if currPublishedVersion != nil { // need handle previous published version
				// update previous published version to draft
				err = u.articleRepo.UpdateArticleVersionStatus(tx, &entity.UpdateArticleVersionStatusRequest{
					ArticleSerial: req.ArticleSerial,
					VersionSerial: currPublishedVersion.Serial,
					NewStatus:     entity.VersionStatusDraft.String(),
				})
				if err != nil {
					return err
				}

				// decrement tag usage count the previous pubslihed version
				currPublishedVersionTagSerials := currPublishedVersion.TagSerials()
				allAffectedTagSerials = append(allAffectedTagSerials, currPublishedVersionTagSerials...)

				err = u.tagRepo.DecrementUsageCount(tx, currPublishedVersionTagSerials)
				if err != nil {
					return err
				}
			}

			// increment tag usage count for new published version
			err = u.tagRepo.IncrementUsageCount(tx, tagsSerials)
			if err != nil {
				return err
			}
		} else if entity.IsPublishedStatus(currStatus) && !entity.IsPublishedStatus(newStatus) { // unpublish
			// decrement tag usage count this version
			err = u.tagRepo.DecrementUsageCount(tx, tagsSerials)
			if err != nil {
				return err
			}
		}
	}

	allAffectedTagSerials = generalutil.SanitizeDuplicateSerials(allAffectedTagSerials)
	err = u.updateTrendingScore(tx, allAffectedTagSerials)
	if err != nil {
		return err
	}

	// update to new status
	err = u.articleRepo.UpdateArticleVersionStatus(tx, req)
	if err != nil {
		return err
	}

	return err
}

func (u *articleUsecase) GetArticles(ctx *gin.Context, req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	req.Status = entity.VersionStatusPublished.String()

	resp, err := u.articleRepo.GetArticles(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (u *articleUsecase) GetArticleLatestDetail(articleSerial string) (*entity.GetArticleLatestDetailResponse, error) {
	if articleSerial == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get article latest detail: article serial is mandatory"))
	}

	versionDetails, err := u.articleRepo.GetArticleLatestDetail(articleSerial)
	if err != nil {
		return nil, err
	}

	resp := &entity.GetArticleLatestDetailResponse{}

	switch len(versionDetails) {
	case 0:
		// no versions found
	case 1:
		resp.PublishedVersion = versionDetails[0]
	case 2:
		resp.PublishedVersion = versionDetails[0]
		resp.LatestVersion = versionDetails[1]
	default:
		// only take the first two
		resp.PublishedVersion = versionDetails[0]
		resp.LatestVersion = versionDetails[1]
	}

	return resp, nil
}

func (u *articleUsecase) CreateArticleVersion(ctx *gin.Context, req *entity.CreateArticleVersionRequest) (resp *entity.CreateArticleVersionResponse, err error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	authorUsername := entity.GetContextUsername(ctx)
	if authorUsername == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error create article version: user id not found in context"))
	}

	latestVersionNumber, err := u.articleRepo.GetLatestVersionNumber(req.ArticleSerial)
	if err != nil {
		return nil, err
	}

	versionSerial, err := serialutil.GenerateId(versionSerialPrefix)
	if err != nil {
		return nil, fmt.Errorf("error create article version: error generate version serial: %s", err.Error())
	}
	version := &entity.Version{
		Serial:         versionSerial,
		AuthorUsername: authorUsername,
		VersionNumber:  latestVersionNumber + 1,
		ArticleSerial:  req.ArticleSerial,
		Title:          req.Title,
		Content:        req.Content,
		Status:         entity.VersionStatusDraft.String(),
	}

	tx, err := u.articleRepo.GetDb().Begin()
	if err != nil {
		return nil, fmt.Errorf("error create article version: failed to begin transaction: %s", err.Error())
	}

	defer func() {
		err = transactionutil.SettleTransaction(tx, err)
	}()

	err = u.articleRepo.InsertVersionTx(tx, version)
	if err != nil {
		return
	}

	err = u.articleRepo.InsertVersionTagsTx(tx, versionSerial, req.TagSerials)
	if err != nil {
		return
	}

	resp = &entity.CreateArticleVersionResponse{
		ArticleSerial: req.ArticleSerial,
		AuthorId:      authorUsername,
		Version:       version,
	}

	return
}

func (u *articleUsecase) DeleteArticle(articleSerial string) error {
	if articleSerial == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error delete article: article serial is mandatory"))
	}

	var currPublishedVersion *entity.Version
	publishedVersions, err := u.articleRepo.GetVersionsByQuery(&entity.GetVersionsByQueryRequest{
		ArticleSerial: articleSerial,
		Status:        entity.VersionStatusPublished.String(),
	})
	if err != nil {
		return fmt.Errorf("error get published version: %s", err.Error())
	}
	if len(publishedVersions) > 0 {
		currPublishedVersion = publishedVersions[0]
	}

	tx := u.transactionPkg.InitTransaction()
	defer func() {
		u.transactionPkg.SettleTransaction(tx, err)
	}()

	err = u.articleRepo.DeleteArticle(tx, articleSerial)
	if err != nil {
		return err
	}

	err = u.articleRepo.DeleteVersionByArticleSerial(tx, articleSerial)
	if err != nil {
		return err
	}

	if currPublishedVersion != nil {
		currPublishedVersionTagSerials := currPublishedVersion.TagSerials()
		// decrement tag usage count the previous pubslihed version
		err = u.tagRepo.DecrementUsageCount(tx, currPublishedVersionTagSerials)
		if err != nil {
			return err
		}

		err = u.updateTrendingScore(tx, currPublishedVersion.TagSerials())
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *articleUsecase) GetVersionsByArticleSerial(articleSerial string) (*entity.GetVersionsByArticleSerialResponse, error) {
	if articleSerial == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get versions by article serial: article serial is mandatory"))
	}

	versions, err := u.articleRepo.GetVersionsByQuery(&entity.GetVersionsByQueryRequest{
		ArticleSerial: articleSerial,
	})
	if err != nil {
		return nil, err
	}

	return &entity.GetVersionsByArticleSerialResponse{
		Versions: versions,
	}, nil
}

func (u *articleUsecase) GetVersionBySerial(serial string) (*entity.Version, error) {
	if serial == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get version by serial: serial is mandatory"))
	}

	return u.articleRepo.GetVersionBySerial(serial)
}

func (u *articleUsecase) calculateTrendingScore(usageCount int, lastUpdatedAt time.Time) float32 {
	if usageCount <= 0 {
		return 0
	}

	// age in days
	ageDays := time.Since(lastUpdatedAt).Hours() / 24

	// decay rate from half-life
	lambda := math.Ln2 / u.cfg.TrendingScoreHalLifeDays

	// decay formula
	recencyFactor := math.Exp(-float64(lambda) * ageDays)

	return float32(usageCount) * float32(recencyFactor)
}

func (u *articleUsecase) updateTrendingScore(tx *gorm.DB, tagSerials []string) error {
	tagStats, err := u.tagRepo.GetTagStatsBySerials(tx, tagSerials)
	if err != nil {
		return err
	}

	for _, tagStat := range tagStats {
		newTrendingScore := u.calculateTrendingScore(int(tagStat.UsageCount), tagStat.UpdatedAt)
		err = u.tagRepo.UpdateTagStat(tx, tagStat.TagSerial, newTrendingScore)
		if err != nil {
			return err
		}
	}

	return nil
}
