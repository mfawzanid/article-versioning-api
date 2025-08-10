package usecase

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	serialutil "article-versioning-api/utils/serial"
	transactionutil "article-versioning-api/utils/transaction"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type articleUsecase struct {
	articleRepo repository.ArticleRepositoryInterface
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

func NewArticleUsecase(articleRepo repository.ArticleRepositoryInterface) ArticleUsecaseInterface {
	return &articleUsecase{articleRepo}
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

func (u *articleUsecase) UpdateArticleVersionStatus(req *entity.UpdateArticleVersionStatusRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return u.articleRepo.UpdateArticleVersionStatus(req)
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

	tx, err := u.articleRepo.GetDb().Begin()
	if err != nil {
		return fmt.Errorf("error delete article: failed to begin transaction: %s", err.Error())
	}

	defer func() {
		err = transactionutil.SettleTransaction(tx, err)
	}()

	err = u.articleRepo.DeleteArticle(tx, articleSerial)
	if err != nil {
		return err
	}

	err = u.articleRepo.DeleteVersionByArticleSerial(tx, articleSerial)
	if err != nil {
		return err
	}

	return nil
}

func (u *articleUsecase) GetVersionsByArticleSerial(articleSerial string) (*entity.GetVersionsByArticleSerialResponse, error) {
	if articleSerial == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get versions by article serial: article serial is mandatory"))
	}

	versions, err := u.articleRepo.GetVersionsByArticleSerial(articleSerial)
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
