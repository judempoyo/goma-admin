package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/models"
	"github.com/jkaninda/goma-admin/store"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	users                store.UserStore
	roles                store.RoleStore
	refreshTokens        store.RefreshTokenStore
	oauth                store.OAuthStore
	tokens               *TokenService
	passwordPolicy       PasswordPolicy
	allowFirstAdmin      bool
	requireVerifiedEmail bool
	state                *StateService
	providers            map[string]OAuthProvider
	now                  func() time.Time
}

type ServiceConfig struct {
	Users                store.UserStore
	Roles                store.RoleStore
	RefreshTokens        store.RefreshTokenStore
	OAuth                store.OAuthStore
	Tokens               *TokenService
	PasswordPolicy       PasswordPolicy
	AllowFirstAdmin      bool
	RequireVerifiedEmail bool
	State                *StateService
	Providers            map[string]OAuthProvider
}

func NewService(cfg ServiceConfig) *Service {
	return &Service{
		users:                cfg.Users,
		roles:                cfg.Roles,
		refreshTokens:        cfg.RefreshTokens,
		oauth:                cfg.OAuth,
		tokens:               cfg.Tokens,
		passwordPolicy:       cfg.PasswordPolicy,
		allowFirstAdmin:      cfg.AllowFirstAdmin,
		requireVerifiedEmail: cfg.RequireVerifiedEmail,
		state:                cfg.State,
		providers:            cfg.Providers,
		now:                  time.Now,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if err := s.passwordPolicy.Validate(email, input.Password); err != nil {
		return nil, err
	}
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:            uuid.New(),
		Email:         email,
		Name:          strings.TrimSpace(input.Name),
		PasswordHash:  string(hashed),
		EmailVerified: false,
	}

	roleNames := []string{"user"}
	if s.allowFirstAdmin {
		count, err := s.users.Count(ctx)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			roleNames = []string{"admin"}
		}
	}
	roles, err := s.roles.EnsureRoles(ctx, roleNames)
	if err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}
	if len(roles) > 0 {
		if err := s.users.AddRoles(ctx, user, roles); err != nil {
			return nil, err
		}
		user.Roles = roles
	}
	return s.issueTokens(ctx, user, roles)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.PasswordHash == "" {
		return nil, ErrPasswordLoginDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user, user.Roles)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, ErrRefreshTokenInvalid
	}
	hash := HashRefreshToken(refreshToken)
	stored, err := s.refreshTokens.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	if stored.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}
	if stored.ExpiresAt.Before(s.now()) {
		return nil, ErrRefreshTokenExpired
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID, s.now()); err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, user, user.Roles)
}

func (s *Service) Logout(ctx context.Context, refreshToken string, all bool) error {
	if strings.TrimSpace(refreshToken) == "" {
		return ErrRefreshTokenInvalid
	}
	hash := HashRefreshToken(refreshToken)
	stored, err := s.refreshTokens.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRefreshTokenInvalid
		}
		return err
	}
	if all {
		return s.refreshTokens.RevokeAllForUser(ctx, stored.UserID, s.now())
	}
	return s.refreshTokens.Revoke(ctx, stored.ID, s.now())
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *Service) OAuthStart(providerName, redirect string) (string, error) {
	provider := s.provider(providerName)
	if provider == nil {
		return "", ErrOAuthProviderNotConfigured
	}
	state, err := s.state.New(provider.Name(), redirect)
	if err != nil {
		return "", err
	}
	return provider.AuthCodeURL(state), nil
}

func (s *Service) OAuthCallback(ctx context.Context, providerName, code, state string) (*AuthResult, error) {
	provider := s.provider(providerName)
	if provider == nil {
		return nil, ErrOAuthProviderNotConfigured
	}
	claims, err := s.state.Parse(state)
	if err != nil || claims.Provider != provider.Name() {
		return nil, ErrOAuthStateInvalid
	}

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	profile, err := provider.Profile(ctx, token)
	if err != nil {
		return nil, err
	}
	if profile.Email == "" {
		return nil, ErrOAuthEmailMissing
	}
	if s.requireVerifiedEmail && !profile.EmailVerified {
		return nil, ErrOAuthEmailNotVerified
	}

	account, err := s.oauth.FindByProviderID(ctx, provider.Name(), profile.ProviderUserID)
	switch {
	case err == nil:
		user, err := s.users.FindByID(ctx, account.UserID)
		if err != nil {
			return nil, err
		}
		expiresAt := expiresAtFromToken(s.now(), token.ExpiresIn)
		_ = s.oauth.UpdateTokens(ctx, account.ID, token.AccessToken, token.RefreshToken, expiresAt)
		return s.issueTokens(ctx, user, user.Roles)
	case errors.Is(err, gorm.ErrRecordNotFound):
		// continue
	default:
		return nil, err
	}

	userEmail := strings.ToLower(strings.TrimSpace(profile.Email))
	user, err := s.users.FindByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = &models.User{
				ID:            uuid.New(),
				Email:         userEmail,
				Name:          strings.TrimSpace(profile.Name),
				PasswordHash:  "",
				EmailVerified: profile.EmailVerified,
			}
			roleNames := []string{"user"}
			if s.allowFirstAdmin {
				count, err := s.users.Count(ctx)
				if err != nil {
					return nil, err
				}
				if count == 0 {
					roleNames = []string{"admin"}
				}
			}
			roles, err := s.roles.EnsureRoles(ctx, roleNames)
			if err != nil {
				return nil, err
			}
			if err := s.users.Create(ctx, user); err != nil {
				return nil, err
			}
			if len(roles) > 0 {
				if err := s.users.AddRoles(ctx, user, roles); err != nil {
					return nil, err
				}
				user.Roles = roles
			}
		} else {
			return nil, err
		}
	}

	account = &models.OAuthAccount{
		Provider:       provider.Name(),
		ProviderUserID: profile.ProviderUserID,
		UserID:         user.ID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresAt:      expiresAtFromToken(s.now(), token.ExpiresIn),
	}
	if err := s.oauth.Create(ctx, account); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, user.Roles)
}

func (s *Service) issueTokens(ctx context.Context, user *models.User, roles []models.Role) (*AuthResult, error) {
	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}
	accessToken, expiresAt, err := s.tokens.NewAccessToken(user, roleNames)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshHash, refreshExpiresAt, err := s.tokens.NewRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Create(ctx, &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, err
	}
	return &AuthResult{
		AccessToken:          accessToken,
		RefreshToken:         refreshToken,
		AccessTokenExpiresAt: expiresAt,
		User:                 user,
		Roles:                roleNames,
	}, nil
}

func (s *Service) provider(name string) OAuthProvider {
	if s.providers == nil {
		return nil
	}
	return s.providers[strings.ToLower(strings.TrimSpace(name))]
}

func normalizeEmail(email string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(email))
	if trimmed == "" {
		return "", ErrEmailRequired
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return "", ErrInvalidEmail
	}
	return trimmed, nil
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func expiresAtFromToken(now time.Time, expiresIn int64) *time.Time {
	if expiresIn <= 0 {
		return nil
	}
	expiresAt := now.Add(time.Duration(expiresIn) * time.Second)
	return &expiresAt
}
