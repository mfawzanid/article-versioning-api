package entity

import (
	errorutil "article-versioning-api/utils/error"
	"fmt"
)

const (
	ContextUsername = "username"
	ContextRole     = "role"
)

type RegisterUserRequest struct {
	Username string
	Password string
	Hash     string
	Role     string
}

func (r *RegisterUserRequest) Validate() error {
	if r.Username == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error register user request: username is mandatory"))
	}
	if r.Password == "" {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error register user request: username is mandatory"))
	}
	if StringToUserRole(r.Role) == UserRoleUnknown {
		return errorutil.NewCustomError(errorutil.ErrBadRequest, fmt.Errorf("error register user request: user role is not valid"))
	}

	return nil
}

type UserRole int

const (
	UserRoleUnknown UserRole = iota
	UserRoleAdmin
	UserRoleWriter
	UserRoleEditor
)

var (
	mapUserRoleToString = map[UserRole]string{
		UserRoleAdmin:  "admin",
		UserRoleWriter: "writer",
		UserRoleEditor: "editor",
	}
	mapStringToUserRole = map[string]UserRole{
		"admin":  UserRoleAdmin,
		"writer": UserRoleWriter,
		"editor": UserRoleEditor,
	}
)

func (ur UserRole) String() string {
	return mapUserRoleToString[ur]
}

func StringToUserRole(ur string) UserRole {
	userRole, ok := mapStringToUserRole[ur]
	if !ok {
		return UserRoleUnknown
	}

	return userRole
}

type User struct {
	Username string
	Role     string
	Hash     string
}

type LoginRequest struct {
	Username string
	Password string
}
