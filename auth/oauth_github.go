package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitHubProvider struct {
	client       *http.Client
	clientID     string
	clientSecret string
	redirectURL  string
	authURL      string
	tokenURL     string
	userURL      string
	emailsURL    string
}

func NewGitHubProvider(clientID, clientSecret, redirectURL, authURL, tokenURL, userURL, emailsURL string) *GitHubProvider {
	return &GitHubProvider{
		client:       &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		authURL:      authURL,
		tokenURL:     tokenURL,
		userURL:      userURL,
		emailsURL:    emailsURL,
	}
}

func (p *GitHubProvider) Name() string {
	return "github"
}

func (p *GitHubProvider) AuthCodeURL(state string) string {
	query := url.Values{}
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.redirectURL)
	query.Set("scope", "read:user user:email")
	query.Set("state", state)
	return fmt.Sprintf("%s?%s", p.authURL, query.Encode())
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string) (OAuthToken, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("redirect_uri", p.redirectURL)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return OAuthToken{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return OAuthToken{}, fmt.Errorf("github token exchange failed: %s", resp.Status)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return OAuthToken{}, err
	}
	return OAuthToken{
		AccessToken: payload.AccessToken,
		TokenType:   payload.TokenType,
	}, nil
}

func (p *GitHubProvider) Profile(ctx context.Context, token OAuthToken) (OAuthProfile, error) {
	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userURL, nil)
	if err != nil {
		return OAuthProfile{}, err
	}
	userReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github+json")

	userResp, err := p.client.Do(userReq)
	if err != nil {
		return OAuthProfile{}, err
	}
	defer userResp.Body.Close()

	if userResp.StatusCode < http.StatusOK || userResp.StatusCode >= http.StatusMultipleChoices {
		return OAuthProfile{}, fmt.Errorf("github userinfo failed: %s", userResp.Status)
	}

	var userPayload struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Login  string `json:"login"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userPayload); err != nil {
		return OAuthProfile{}, err
	}

	email := ""
	verified := false

	emailsReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.emailsURL, nil)
	if err != nil {
		return OAuthProfile{}, err
	}
	emailsReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	emailsReq.Header.Set("Accept", "application/vnd.github+json")

	emailsResp, err := p.client.Do(emailsReq)
	if err != nil {
		if userPayload.Email == "" {
			return OAuthProfile{}, err
		}
	} else {
		defer emailsResp.Body.Close()
		if emailsResp.StatusCode < http.StatusOK || emailsResp.StatusCode >= http.StatusMultipleChoices {
			if userPayload.Email == "" {
				return OAuthProfile{}, fmt.Errorf("github emails failed: %s", emailsResp.Status)
			}
		} else {
			var emailsPayload []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(emailsResp.Body).Decode(&emailsPayload); err != nil {
				return OAuthProfile{}, err
			}
			for _, item := range emailsPayload {
				if item.Primary {
					email = item.Email
					verified = item.Verified
					break
				}
			}
		}
	}

	if email == "" {
		email = userPayload.Email
	}

	name := userPayload.Name
	if name == "" {
		name = userPayload.Login
	}

	return OAuthProfile{
		Provider:       p.Name(),
		ProviderUserID: fmt.Sprintf("%d", userPayload.ID),
		Email:          email,
		EmailVerified:  verified,
		Name:           name,
		AvatarURL:      userPayload.Avatar,
	}, nil
}
