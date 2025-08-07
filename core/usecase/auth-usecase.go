package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"

	"article-versioning-api/core/entity"
	errorutil "article-versioning-api/utils/error"
)

type AuthUsecaseInterface interface {
	CreateToken(user *entity.User) (tokenString string, err error)
	VerifyToken(tokenString string) (*entity.User, error)
}

type authUsecase struct {
}

func NewAuthUsecase() AuthUsecaseInterface {
	return &authUsecase{}
}

const (
	tokenSecret = "token_secret" // TODO: set in config
)

func (u *authUsecase) CreateToken(user *entity.User) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		entity.ContextUsername: user.Username,
		entity.ContextRole:     user.Role,
		"exp":                  time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err = token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error create token: %v", err.Error())
	}

	return tokenString, nil
}

func (u *authUsecase) VerifyToken(tokenString string) (*entity.User, error) {
	token, err := jwt.Parse(tokenString, func(tokenString *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return nil, errorutil.NewCustomError(errorutil.ErrUnauthorized, fmt.Errorf("error verify token: %v", err.Error()))
	}
	if !token.Valid {
		return nil, errorutil.NewCustomError(errorutil.ErrUnauthorized, errors.New("error verify token: token is invalid"))
	}

	username, ok := token.Claims.(jwt.MapClaims)[entity.ContextUsername].(string)
	if !ok {
		return nil, errorutil.NewCustomError(errorutil.ErrUnauthorized, errors.New("error verify token: user name is not found in token"))
	}
	role, ok := token.Claims.(jwt.MapClaims)[entity.ContextRole].(string)
	if !ok {
		return nil, errorutil.NewCustomError(errorutil.ErrUnauthorized, errors.New("error verify token: role is not found in token"))
	}

	return &entity.User{
		Username: username,
		Role:     role,
	}, nil
}
