package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/models"
)

type AccessTokenClaims struct {
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret          []byte
	issuer          string
	audience        string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	now             func() time.Time
}

func NewTokenService(secret, issuer, audience string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:          []byte(secret),
		issuer:          issuer,
		audience:        audience,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		now:             time.Now,
	}
}

func (s *TokenService) NewAccessToken(user *models.User, roles []string) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, fmt.Errorf("user is nil")
	}
	expiresAt := s.now().Add(s.accessTokenTTL)
	claims := AccessTokenClaims{
		Email: user.Email,
		Name:  user.Name,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    s.issuer,
			Audience:  []string{s.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(s.now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (s *TokenService) NewRefreshToken(userID uuid.UUID) (string, string, time.Time, error) {
	plain, err := generateRandomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}
	hash := HashRefreshToken(plain)
	expiresAt := s.now().Add(s.refreshTokenTTL)
	return plain, hash, expiresAt, nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func generateRandomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
