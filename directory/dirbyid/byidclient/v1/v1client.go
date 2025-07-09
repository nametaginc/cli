package v1

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2/clientcredentials"
)

// V1Client is the client for the Beyond Identity v1 API.
type V1Client struct {
	baseURL  *url.URL
	tenantID string
	realmID  string

	client *http.Client
}

// NewV1Client creates a new V1Client.
func NewV1Client(apiBaseURL *url.URL, clientID, clientSecret, tenantID, realmID, applicationID string) (*V1Client, error) {
	// If api BaseURL is https://api-us.beyondidentity.com/v1,
	// then tokenBaseURL is https://auth-us.beyondidentity.com/v1
	tokenBaseURL := strings.Replace(apiBaseURL.String(), "api", "auth", 1)

	// tokenURL is https://auth-us.beyondidentity.com/v1/tenants/$TENANT_ID/realms/$REALM_ID/applications/$APPLICATION_ID/token.
	// See https://docs.beyondidentity.com/api/v1#section/Authentication.
	tokenURL, err := url.JoinPath(tokenBaseURL, "v1", "tenants", tenantID, "realms", realmID, "applications", applicationID, "token")
	if err != nil {
		return nil, err
	}

	log.Printf("tokenURL: %s", tokenURL)

	cfg := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{"identities:read", "groups:read"},
	}

	// This httpClient automatically adds Authorization header and handles token expiration.
	httpClient := cfg.Client(context.Background())

	return &V1Client{
		baseURL:  apiBaseURL,
		tenantID: tenantID,
		realmID:  realmID,
		client:   httpClient,
	}, nil
}
