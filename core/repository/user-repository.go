package repository

import "article-versioning-api/core/entity"

type UserRepositoryInterface interface {
	CreateUser(req *entity.RegisterUserRequest) error
	GetUserByUsername(username string) (*entity.User, error)
}
