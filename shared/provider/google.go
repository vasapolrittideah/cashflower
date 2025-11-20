package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var (
	ErrInvalidGoogleAudience = errors.New("invalid google audience")
)

type GoogleOAuthProvider struct {
	idToken  string
	clientID string
}

func (p *GoogleOAuthProvider) ValidateIDToken(ctx context.Context) (*oauth2.Tokeninfo, error) {
	oauth2Service, err := oauth2.NewService(ctx, option.WithHTTPClient(&http.Client{}))
	if err != nil {
		return nil, err
	}

	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(p.idToken)
	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}

	if tokenInfo.Audience != p.clientID {
		return nil, ErrInvalidGoogleAudience
	}

	return tokenInfo, nil
}

func (p *GoogleOAuthProvider) GetUserInfo() (*oauth2.Userinfo, error) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.idToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status code is not OK")
	}

	var userInfo oauth2.Userinfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
