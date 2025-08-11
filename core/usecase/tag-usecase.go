package usecase

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	serialutil "article-versioning-api/utils/serial"
	transactionutil "article-versioning-api/utils/transaction"
	"errors"
	"fmt"
)

type tagUsecase struct {
	tagRepo repository.TagRepositoryInterface
	transactionPkg        transactionutil.Transaction
}

type TagUsecaseInterface interface {
	CreateTag(req *entity.CreateTagRequest) (serial string, err error)
	GetTags(req *entity.GetTagsRequest) (*entity.GetTagsResponse, error)
	GetTagBySerial(serial string) (*entity.TagDetail, error)
}

func NewTagUsecase(tagRepo repository.TagRepositoryInterface, transactionPkg transactionutil.Transaction) TagUsecaseInterface {
	return &tagUsecase{tagRepo, transactionPkg}
}

const (
	tagSerialPrefix = "TAG"
)

func (u *tagUsecase) CreateTag(req *entity.CreateTagRequest) (serial string, err error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	serial, err = serialutil.GenerateId(tagSerialPrefix)
	if err != nil {
		return "", fmt.Errorf("error create tag: error generate serial: %s", err.Error())
	}

	tx := u.transactionPkg.InitTransaction()
	defer u.transactionPkg.SettleTransaction(tx, err)

	err = u.tagRepo.InsertTag(&entity.Tag{
		Serial: serial,
		Name:   req.Name,
	}, tx)
	if err != nil {
		return "", err
	}

	err = u.tagRepo.InsertTagStat(serial, tx)
	if err != nil {
		return "", err
	}

	return serial, nil
}

func (u *tagUsecase) GetTags(req *entity.GetTagsRequest) (*entity.GetTagsResponse, error) {
	if req.Pagination != nil {
		req.Pagination.Validate()
	}

	tagDetails, err := u.tagRepo.GetTags(req.Pagination)
	if err != nil {
		return nil, err
	}

	return &entity.GetTagsResponse{
		Tags:       tagDetails,
		Pagination: req.Pagination,
	}, nil
}

func (u *tagUsecase) GetTagBySerial(serial string) (*entity.TagDetail, error) {
	if serial == "" {
		return nil, errorutil.NewCustomError(errorutil.ErrBadRequest, errors.New("error get tag by serial: serial is mandatory"))
	}

	return u.tagRepo.GetTagBySerial(serial)
}
