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
// See the License for the specific governing permissions and
// limitations under the License.

// Package dirbyid impements a Beyond Identity directory provider.
package dirbyid

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
	v0 "github.com/nametaginc/cli/directory/dirbyid/byidclient/v0"
	v1 "github.com/nametaginc/cli/directory/dirbyid/byidclient/v1"
	"github.com/samber/lo"
)

type Provider struct {
	Version       string
	APIBaseURL    *url.URL
	ClientID      string
	ClientSecret  string
	TenantID      *string
	RealmID       *string
	ApplicationID *string

	// Internal client for Beyond Identity API.
	// Agnostic of the v0 or v1 API.
	client byidclient.Client
}

func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	err := p.initClient()
	if err != nil {
		return nil, err
	}

	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                    "Beyond Identity",
			CanGetTemporaryPassword: lo.ToPtr(false),
			CanGetPasswordLink:      lo.ToPtr(false),
			CanRemoveAllMFA:         lo.ToPtr(false),
			CanUnlock:               lo.ToPtr(false),
			CanUpdateAccountsList:   lo.ToPtr(false),
		},
		ImmutableID: fmt.Sprintf("urn:agent:%s", p.ClientID),
	}, nil
}

// initClient initializes the client for the Beyond Identity API.
//
// If the TenantID and RealmID are provided, it uses the v1 API.
// Otherwise, it uses the v0 API.
func (p *Provider) initClient() error {
	var err error
	if p.TenantID != nil && p.RealmID != nil {
		p.client, err = v1.NewV1Client(p.APIBaseURL, p.ClientID, p.ClientSecret, *p.TenantID, *p.RealmID, *p.ApplicationID)
	} else {
		p.client, err = v0.NewV0Client(p.APIBaseURL, p.ClientID, p.ClientSecret)
	}
	if err != nil {
		return err
	}
	return nil
}
