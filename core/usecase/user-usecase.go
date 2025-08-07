package usecase

import (
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	errorutil "article-versioning-api/utils/error"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepository repository.UserRepositoryInterface
	authUsecase    AuthUsecaseInterface
}

type UserUsecaseInterface interface {
	RegisterUser(req *entity.RegisterUserRequest) error
	Login(req *entity.LoginRequest) (token string, err error)
}

func NewUserUsecase(userRepository repository.UserRepositoryInterface, authUsecase AuthUsecaseInterface) UserUsecaseInterface {
	return &userUsecase{userRepository, authUsecase}
}

func (u *userUsecase) RegisterUser(req *entity.RegisterUserRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hash, err := generateHashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("error register user: %s", err.Error())
	}
	req.Hash = hash

	return u.userRepository.CreateUser(req)
}

func (u *userUsecase) Login(req *entity.LoginRequest) (token string, err error) {
	user, err := u.userRepository.GetUserByUsername(req.Username)
	if err != nil {
		return "", err
	}

	err = validatePassword(user.Hash, req.Password)
	if err != nil {
		return "", errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error login: %s", err.Error()))
	}

	token, err = u.authUsecase.CreateToken(user)
	if err != nil {
		return token, err
	}

	return token, nil
}

func generateHashPassword(password string) (string, error) {
	bytePassword := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func validatePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
