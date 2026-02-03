package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jkaninda/goma-admin/models"
)

var (
	ErrEmailRequired              = errors.New("email is required")
	ErrInvalidEmail               = errors.New("invalid email address")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrPasswordLoginDisabled      = errors.New("password login disabled")
	ErrRefreshTokenInvalid        = errors.New("invalid refresh token")
	ErrRefreshTokenExpired        = errors.New("refresh token expired")
	ErrRefreshTokenRevoked        = errors.New("refresh token revoked")
	ErrOAuthProviderNotConfigured = errors.New("oauth provider not configured")
	ErrOAuthEmailMissing          = errors.New("oauth email missing")
	ErrOAuthEmailNotVerified      = errors.New("oauth email not verified")
	ErrOAuthStateInvalid          = errors.New("oauth state invalid")
)

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	AccessToken          string
	RefreshToken         string
	AccessTokenExpiresAt time.Time
	User                 *models.User
	Roles                []string
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type OAuthProfile struct {
	Provider       string
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Name           string
	AvatarURL      string
}

type OAuthProvider interface {
	Name() string
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (OAuthToken, error)
	Profile(ctx context.Context, token OAuthToken) (OAuthProfile, error)
}
