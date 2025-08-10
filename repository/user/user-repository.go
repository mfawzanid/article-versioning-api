package userrepository

import (
	"article-versioning-api/config"
	"article-versioning-api/core/entity"
	"article-versioning-api/core/repository"
	"database/sql"
	"fmt"

	errorutil "article-versioning-api/utils/error"

	"github.com/lib/pq"
)

type userRepository struct {
	Db *sql.DB
	cfg *config.Config
}

func NewUserRepository(Db *sql.DB, cfg *config.Config) repository.UserRepositoryInterface {
	return &userRepository{Db, cfg}
}


func (r *userRepository) CreateUser(req *entity.RegisterUserRequest) error {
	query := `INSERT INTO users (username, role, hash) VALUES ($1, $2, $3)`

	_, err := r.Db.Exec(query, req.Username, req.Role, req.Hash)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pq.ErrorCode(r.cfg.PSQLUniqueViolationErrorCode) {
			return errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error create user: username has exist"))
		} else {
			return fmt.Errorf("error repo create user: %v", err.Error())
		}
	}

	return nil
}

func (r *userRepository) GetUserByUsername(username string) (*entity.User, error) {
	query := `SELECT username, role, hash FROM users WHERE username = $1`

	user := &entity.User{}

	err := r.Db.QueryRow(query, username).Scan(&user.Username, &user.Role, &user.Hash)
	if err != nil {
		return nil, fmt.Errorf("error repo create user: %s", err.Error())
	}

	return user, nil
}
