package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/auth"
	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/goma-admin/models"
	"github.com/jkaninda/goma-admin/store"
	"github.com/jkaninda/okapi"
)

type AuthService struct {
	service *auth.Service
}

func NewAuthService(conf *config.Config) *AuthService {
	userStore := store.NewUserStore(conf.Database.DB)
	roleStore := store.NewRoleStore(conf.Database.DB)
	refreshStore := store.NewRefreshTokenStore(conf.Database.DB)
	oauthStore := store.NewOAuthStore(conf.Database.DB)

	tokenService := auth.NewTokenService(
		conf.JWT.Secret,
		conf.JWT.Issuer,
		conf.JWT.Audience,
		conf.JWT.AccessTokenTTL,
		conf.JWT.RefreshTokenTTL,
	)
	stateService := auth.NewStateService(conf.JWT.Secret, conf.JWT.Issuer, 5*time.Minute)

	providers := map[string]auth.OAuthProvider{}
	if google := conf.Auth.OAuth.Google; isOAuthConfigured(google) {
		providers["google"] = auth.NewGoogleProvider(
			google.ClientID,
			google.ClientSecret,
			google.RedirectURL,
			defaultIfEmpty(google.AuthURL, "https://accounts.google.com/o/oauth2/v2/auth"),
			defaultIfEmpty(google.TokenURL, "https://oauth2.googleapis.com/token"),
			defaultIfEmpty(google.UserInfoURL, "https://openidconnect.googleapis.com/v1/userinfo"),
		)
	}
	if github := conf.Auth.OAuth.GitHub; isOAuthConfigured(github) {
		providers["github"] = auth.NewGitHubProvider(
			github.ClientID,
			github.ClientSecret,
			github.RedirectURL,
			defaultIfEmpty(github.AuthURL, "https://github.com/login/oauth/authorize"),
			defaultIfEmpty(github.TokenURL, "https://github.com/login/oauth/access_token"),
			defaultIfEmpty(github.UserInfoURL, "https://api.github.com/user"),
			defaultIfEmpty(github.EmailsURL, "https://api.github.com/user/emails"),
		)
	}

	service := auth.NewService(auth.ServiceConfig{
		Users:         userStore,
		Roles:         roleStore,
		RefreshTokens: refreshStore,
		OAuth:         oauthStore,
		Tokens:        tokenService,
		PasswordPolicy: auth.PasswordPolicy{
			MinLength:      conf.Auth.PasswordPolicy.MinLength,
			MaxLength:      conf.Auth.PasswordPolicy.MaxLength,
			RequireUpper:   conf.Auth.PasswordPolicy.RequireUpper,
			RequireLower:   conf.Auth.PasswordPolicy.RequireLower,
			RequireNumber:  conf.Auth.PasswordPolicy.RequireNumber,
			RequireSpecial: conf.Auth.PasswordPolicy.RequireSpecial,
		},
		AllowFirstAdmin:      conf.Auth.AllowFirstAdmin,
		RequireVerifiedEmail: conf.Auth.RequireVerifiedEmail,
		State:                stateService,
		Providers:            providers,
	})
	return &AuthService{service: service}
}

func (s *AuthService) Register(c *okapi.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	result, err := s.service.Register(c.Request().Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Created(toAuthResponse(result))
}

func (s *AuthService) Login(c *okapi.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	result, err := s.service.Login(c.Request().Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.OK(toAuthResponse(result))
}

func (s *AuthService) Refresh(c *okapi.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	result, err := s.service.Refresh(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.OK(toAuthResponse(result))
}

func (s *AuthService) Logout(c *okapi.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	if err := s.service.Logout(c.Request().Context(), req.RefreshToken, req.All); err != nil {
		return handleAuthError(c, err)
	}
	return c.OK(okapi.M{"status": "ok"})
}

func (s *AuthService) Me(c *okapi.Context) error {
	userID := c.GetString("user_id")
	if userID == "" {
		return c.AbortUnauthorized("Missing user id")
	}
	parsed, err := uuid.Parse(userID)
	if err != nil {
		return c.AbortUnauthorized("Invalid user id", err)
	}
	user, err := s.service.Me(c.Request().Context(), parsed)
	if err != nil {
		return c.AbortInternalServerError("Failed to load user", err)
	}
	roles := extractRoleNames(user)
	return c.OK(UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Roles: roles,
	})
}

func (s *AuthService) OAuthStart(c *okapi.Context) error {
	provider := c.Param("provider")
	redirect := c.Query("redirect")
	url, err := s.service.OAuthStart(provider, redirect)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.OK(OAuthStartResponse{URL: url})
}

func (s *AuthService) OAuthCallback(c *okapi.Context) error {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return c.AbortBadRequest("Missing code or state")
	}
	result, err := s.service.OAuthCallback(c.Request().Context(), provider, code, state)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.OK(toAuthResponse(result))
}

func toAuthResponse(result *auth.AuthResult) AuthResponse {
	return AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.AccessTokenExpiresAt.Unix(),
		TokenType:    "Bearer",
		User: UserResponse{
			ID:    result.User.ID.String(),
			Email: result.User.Email,
			Name:  result.User.Name,
			Roles: result.Roles,
		},
	}
}

func handleAuthError(c *okapi.Context, err error) error {
	var policyErr auth.PasswordPolicyError
	switch {
	case errors.Is(err, auth.ErrEmailRequired),
		errors.Is(err, auth.ErrInvalidEmail):
		return c.AbortBadRequest(err.Error(), err)
	case errors.As(err, &policyErr):
		return c.AbortBadRequest(policyErr.Error(), err)
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		return c.AbortConflict("Email already exists", err)
	case errors.Is(err, auth.ErrInvalidCredentials):
		return c.AbortUnauthorized("Invalid credentials", err)
	case errors.Is(err, auth.ErrPasswordLoginDisabled):
		return c.AbortUnauthorized("Password login disabled", err)
	case errors.Is(err, auth.ErrRefreshTokenInvalid),
		errors.Is(err, auth.ErrRefreshTokenExpired),
		errors.Is(err, auth.ErrRefreshTokenRevoked):
		return c.AbortUnauthorized("Invalid refresh token", err)
	case errors.Is(err, auth.ErrOAuthProviderNotConfigured):
		return c.AbortBadRequest("OAuth provider not configured", err)
	case errors.Is(err, auth.ErrOAuthEmailMissing):
		return c.AbortBadRequest("OAuth provider did not return an email", err)
	case errors.Is(err, auth.ErrOAuthEmailNotVerified):
		return c.AbortForbidden("OAuth email is not verified", err)
	case errors.Is(err, auth.ErrOAuthStateInvalid):
		return c.AbortBadRequest("Invalid OAuth state", err)
	default:
		return c.AbortInternalServerError("Authentication error", err)
	}
}

func isOAuthConfigured(cfg config.OAuthProviderConfig) bool {
	return strings.TrimSpace(cfg.ClientID) != "" &&
		strings.TrimSpace(cfg.ClientSecret) != "" &&
		strings.TrimSpace(cfg.RedirectURL) != ""
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func extractRoleNames(user *models.User) []string {
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	return roles
}
