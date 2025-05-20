// Copyright 2025 Nametag Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dirokta impements an Okta directory provider.
package dirokta

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/rehttp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// Provider represents the configuration details required to connect to Okta.
type Provider struct {
	URL          string
	Token        string
	ClientID     string
	ClientSecret string

	Client *okta.Client
}

// Configure returns static information about the integration
func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                  "Okta",
			CanGetPasswordLink:    lo.ToPtr(true),
			CanRemoveAllMFA:       lo.ToPtr(true),
			CanUnlock:             lo.ToPtr(true),
			CanUpdateAccountsList: lo.ToPtr(true),
		},
		ImmutableID: fmt.Sprintf("urn:agent:%s", p.URL),
	}, nil
}

func (p *Provider) client(ctx context.Context) (context.Context, *okta.Client, error) {
	if p.Client != nil {
		return ctx, p.Client, nil
	}
	if p.Token != "" {
		httpClient := &http.Client{
			Transport: retry(http.DefaultTransport),
		}

		ctx, client, err := okta.NewClient(ctx,
			okta.WithHttpClientPtr(httpClient),
			okta.WithOrgUrl(p.URL),
			okta.WithToken(p.Token),
			okta.WithRequestTimeout(120),
			okta.WithRateLimitMaxRetries(10))
		if err != nil {
			return nil, nil, fmt.Errorf("cannot initialize okta client: %w", err)
		}
		p.Client = client
		return ctx, client, nil
	}

	if p.ClientID != "" && p.ClientSecret != "" {
		makeClientAssertion := p.makeClientAssertion(p.ClientID, p.ClientSecret)

		clientAssertion, _, err := makeClientAssertion()
		if err != nil {
			return nil, nil, err
		}

		ctx, client, err := okta.NewClient(ctx,
			okta.WithOrgUrl(p.URL),
			okta.WithAuthorizationMode("JWT"),
			okta.WithClientAssertion(clientAssertion),
			okta.WithScopes(oktaScopes),
			okta.WithRequestTimeout(120),
			okta.WithRateLimitMaxRetries(10))
		if err != nil {
			return nil, nil, err
		}

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(oktaClientAssertionTTL / 2):
					clientAssertion, _, err := makeClientAssertion()
					if err != nil {
						log.Printf("ERROR: failed to refresh okta client assertion: %v", err)
					} else {
						if err := client.SetConfig(okta.WithClientAssertion(clientAssertion)); err != nil {
							log.Printf("ERROR: failed to refresh okta client assertion: %v", err)
						}
					}
				}
			}
		}()

		p.Client = client
		return ctx, client, nil
	}

	return nil, nil, fmt.Errorf("okta directory credentials not configured")
}

const oktaClientAssertionTTL = time.Hour

func (p *Provider) makeClientAssertion(clientID string, clientSecret string) func() (string, time.Time, error) {
	return func() (string, time.Time, error) {
		expires := time.Now().Add(oktaClientAssertionTTL)
		claims := jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{p.URL + "/oauth2/v1/token"},
			ExpiresAt: jwt.NewNumericDate(expires),
			Issuer:    clientID,
			Subject:   clientID,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte(clientSecret))
		if err != nil {
			return "", time.Time{}, err
		}
		return signedToken, expires, nil
	}
}

func retry(transport http.RoundTripper) http.RoundTripper {
	return rehttp.NewTransport(transport,
		rehttp.RetryAll(
			rehttp.RetryMaxRetries(5),
			rehttp.RetryAny(
				rehttp.RetryTemporaryErr(),
				rehttp.RetryStatusInterval(500, 600),
			),
		),
		rehttp.ExpJitterDelay(100*time.Millisecond, 1*time.Second),
	)
}

var oktaScopes = []string{
	"okta.users.manage", // reset password
	"okta.orgs.read",
	"okta.groups.read",
	"okta.users.read", // list user factors, etc.
}
