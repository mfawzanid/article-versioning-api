package usecase

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	serialutil "article-versioning-api/utils/serial"
	"errors"
	"fmt"
)

type tagUsecase struct {
	tagRepo repository.TagRepositoryInterface
}

type TagUsecaseInterface interface {
	CreateTag(req *entity.CreateTagRequest) (serial string, err error)
	GetTags(req *entity.GetTagsRequest) (*entity.GetTagsResponse, error)
	GetTagBySerial(serial string) (*entity.TagDetail, error)
}

func NewTagUsecase(tagRepo repository.TagRepositoryInterface) TagUsecaseInterface {
	return &tagUsecase{tagRepo}
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

	err = u.tagRepo.InsertTag(&entity.Tag{
		Serial: serial,
		Name:   req.Name,
	})
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
