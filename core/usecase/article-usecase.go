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
	UpdateTrendingScoreTags(pg *entity.Pagination) (err error)
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

	err = u.articleRepo.InsertArticleTx(tx, &entity.Article{
		Serial: articleSerial,
	})
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
	tagsSerials := version.TagSerials()

	if currStatus == newStatus {
		return nil
	}
	if entity.IsPublishedStatus(currStatus) == entity.IsPublishedStatus(newStatus) || !entity.IsPublishedStatus(currStatus) == !entity.IsPublishedStatus(newStatus) {
		// published to published or non published to non published, then do nothing
		return nil
	}

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

	allAffectedTagSerials = generalutil.SanitizeDuplicateSerials(allAffectedTagSerials)
	allTagStats, err := u.tagRepo.GetTagStatsBySerials(tx, allAffectedTagSerials)
	if err != nil {
		return err
	}

	err = u.updateTrendingScore(tx, allTagStats)
	if err != nil {
		return err
	}

	// update to new status
	err = u.articleRepo.UpdateArticleVersionStatus(tx, req)
	if err != nil {
		return err
	}

	// update tag relationship score
	// generate pair and record the increment
	tagSerialPairCombination := generatePairCombination(tagsSerials)
	for _, pair := range tagSerialPairCombination {
		if len(pair) >= 2 {
			err = u.tagRepo.IncrementTagPairStat(tx, pair[0], pair[1])
			if err != nil {
				return err
			}
		}
	}
	// calculate tag relationship score based on tag usage count and its pair that increase and or decrease before
	err = u.updateTagRelationshipScore(tx, version.Serial, tagsSerials)
	if err != nil {
		return err
	}

	return err
}

func (u *articleUsecase) GetArticles(ctx *gin.Context, req *entity.GetArticlesRequest) (*entity.GetArticlesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	userRole := entity.GetContextRole(ctx)
	if userRole == entity.UserRoleReader.String() {
		req.Status = entity.VersionStatusPublished.String()
	}

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
		return fmt.Errorf("error delete article: %s", err.Error())
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

		// update the trending score
		tagStats, err := u.tagRepo.GetTagStatsBySerials(tx, currPublishedVersion.TagSerials())
		if err != nil {
			return err
		}

		err = u.updateTrendingScore(tx, tagStats)
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

// calculate the trending score using exponential decay with half-life set in config as TrendingScoreHalLifeDays
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

func (u *articleUsecase) updateTrendingScore(tx *gorm.DB, tagStats []*entity.TagStat) error {
	for _, tagStat := range tagStats {
		newTrendingScore := u.calculateTrendingScore(int(tagStat.UsageCount), *tagStat.UsageCountUpdatedAt)
		err := u.tagRepo.UpdateTagStat(tx, tagStat.TagSerial, newTrendingScore)
		if err != nil {
			return err
		}
	}

	return nil
}

func generatePairCombination(serials []string) [][]string {
	var pairs [][]string
	n := len(serials)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// ensure the sequence, this used in tag_pair_stats
			serial1 := serials[i]
			serial2 := serials[j]
			if serial1 > serial2 {
				serial1, serial2 = serial2, serial1
			}

			pairs = append(pairs, []string{serial1, serial2})
		}
	}

	return pairs
}

// calculate tag relationship score using Positive Pointwise Mutual Information (PMI)
func calculateTagRelationshipScore(tag1UsageCount, tag2UsageCount, pairUsageCount, totalPublishedArticle int) float32 {
	// avoid divide by zero or invalid log
	if tag1UsageCount == 0 || tag2UsageCount == 0 || pairUsageCount == 0 || totalPublishedArticle == 0 {
		return 0
	}

	// PMI = log2( (C(i,j) * N) / (C(i) * C(j)) )
	pmi := math.Log2(float64(pairUsageCount) * float64(totalPublishedArticle) /
		(float64(tag1UsageCount) * float64(tag2UsageCount)))

	// PMI+ = max(PMI, 0)
	if pmi < 0 {
		return 0
	}

	return float32(pmi)
}

func (u *articleUsecase) getAllTagUsageCount(tx *gorm.DB, tagSerials []string) (map[string]int, map[string]int, error) {
	mapTagUsageCount := make(map[string]int)
	tagStats, err := u.tagRepo.GetTagStatsBySerials(tx, tagSerials)
	if err != nil {
		return nil, nil, err
	}
	for _, ts := range tagStats {
		mapTagUsageCount[ts.TagSerial] = int(ts.UsageCount)
	}

	mapTagPairUsageCount := make(map[string]int)
	tagPairStats, err := u.tagRepo.GetTagPairStatsBySerials(tx, tagSerials)
	if err != nil {
		return nil, nil, err
	}
	for _, tps := range tagPairStats {
		mapTagPairUsageCount[fmt.Sprint(tps.Tag1Serial, "-", tps.Tag2Serial)] = int(tps.UsageCount)
	}

	return mapTagUsageCount, mapTagPairUsageCount, nil
}

func (u *articleUsecase) updateTagRelationshipScore(tx *gorm.DB, versionSerial string, tagSerials []string) error {
	if len(tagSerials) < 2 {
		return u.articleRepo.UpdateTagRelationshipScore(tx, versionSerial, 0)
	}

	totalPublishedVersion, err := u.articleRepo.GetTotalPublishedArticle(tx)
	if err != nil {
		return err
	}

	mapTagUsageCount, mapTagPairUsageCount, err := u.getAllTagUsageCount(tx, tagSerials)
	if err != nil {
		return err
	}

	tagSerialPairCombination := generatePairCombination(tagSerials)
	var totalScore float32
	for _, pair := range tagSerialPairCombination {
		tag1UsageCount, ok := mapTagUsageCount[pair[0]]
		if !ok {
			err = fmt.Errorf("error update tag relationship score: usage count tag '%s' is not found", pair[0])
			return err
		}
		tag2UsageCount, ok := mapTagUsageCount[pair[1]]
		if !ok {
			err = fmt.Errorf("error update tag relationship score: usage count tag '%s' is not found", pair[0])
			return err
		}
		tagPairSerial := fmt.Sprint(pair[0], "-", pair[1])
		tagPairUsageCount, ok := mapTagPairUsageCount[tagPairSerial]
		if !ok {
			err = fmt.Errorf("error update tag relationship score: usage count tag pair '%s' is not found", tagPairSerial)
			return err
		}

		score := calculateTagRelationshipScore(tag1UsageCount, tag2UsageCount, tagPairUsageCount, totalPublishedVersion)

		totalScore += score
	}

	finalScore := float32(totalScore) / float32(len(tagSerialPairCombination))

	return u.articleRepo.UpdateTagRelationshipScore(tx, versionSerial, finalScore)
}

// update trending score for all tags that triggered by worker
func (u *articleUsecase) UpdateTrendingScoreTags(pg *entity.Pagination) (err error) {
	pg.SetToDefault()

	tx := u.transactionPkg.InitTransaction()
	defer func() {
		u.transactionPkg.SettleTransaction(tx, err)
	}()

	for {
		tagStats, err := u.tagRepo.GetTagStats(tx, pg)
		if err != nil {
			return err
		}

		err = u.updateTrendingScore(tx, tagStats)
		if err != nil {
			return err
		}

		if pg.Page == pg.TotalPage {
			break
		}

		pg.Page++
	}

	return nil
}
