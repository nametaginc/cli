package v0

import (
	"context"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/clientcredentials"
)

// V0Client is the client for the Beyond Identity v0 API.
type V0Client struct {
	baseURL string
	Client  *http.Client
}

// NewV0Client creates a new V0Client.
func NewV0Client(apiBaseURL *url.URL, clientID, clientSecret string) (*V0Client, error) {
	// tokenURL is https://api.byndid.com/v2/oauth2/token.
	// See https://docs.beyondidentity.com/api/v0#section/Authentication.
	tokenURL, err := url.JoinPath(apiBaseURL.String(), "oauth2", "token")
	if err != nil {
		return nil, err
	}

	cfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{"users:read", "groups:read"}, // adjust as needed
	}

	httpClient := cfg.Client(context.Background())

	return &V0Client{
		baseURL: apiBaseURL.String(),
		Client:  httpClient,
	}, nil
}
