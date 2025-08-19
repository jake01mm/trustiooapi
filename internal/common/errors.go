package common

import "errors"

var (
	// 用户相关错误
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrPhoneExists        = errors.New("phone already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrUserInactive       = errors.New("user inactive")

	// 管理员相关错误
	ErrAdminNotFound       = errors.New("admin not found")
	ErrAdminEmailExists    = errors.New("admin email already exists")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrAdminInactive       = errors.New("admin inactive")

	// 验证码相关错误
	ErrCodeNotFound     = errors.New("verification code not found")
	ErrCodeExpired      = errors.New("verification code expired")
	ErrCodeAlreadyUsed  = errors.New("verification code already used")
	ErrInvalidCode      = errors.New("invalid verification code")

	// 令牌相关错误
	ErrTokenNotFound    = errors.New("token not found")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")

	// 权限相关错误
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// 通用错误
	ErrInternalServer   = errors.New("internal server error")
	ErrBadRequest       = errors.New("bad request")
	ErrNotFound         = errors.New("not found")
	ErrValidation       = errors.New("validation error")
)