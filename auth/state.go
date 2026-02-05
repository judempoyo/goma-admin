package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type OAuthStateClaims struct {
	Provider string `json:"provider"`
	Redirect string `json:"redirect,omitempty"`
	jwt.RegisteredClaims
}

type StateService struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

func NewStateService(secret, issuer string, ttl time.Duration) *StateService {
	return &StateService{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *StateService) New(provider, redirect string) (string, error) {
	expiresAt := s.now().Add(s.ttl)
	claims := OAuthStateClaims{
		Provider: provider,
		Redirect: redirect,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "oauth-state",
			Issuer:    s.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(s.now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *StateService) Parse(tokenStr string) (*OAuthStateClaims, error) {
	claims := &OAuthStateClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil || !token.Valid {
		return nil, ErrOAuthStateInvalid
	}
	if claims.Provider == "" {
		return nil, fmt.Errorf("state missing provider")
	}
	return claims, nil
}
