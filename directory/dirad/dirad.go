// Copyright 2024 Nametag Inc.
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

// Package dirad implements an Active Directory agent
package dirad

import (
	"context"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

// Provider represents the directory provider
type Provider struct {
	_client adclient.Client
}

func (p *Provider) client() (adclient.Client, error) {
	if p._client == nil {
		client, err := adclient.New()
		if err != nil {
			return nil, err
		}
		p._client = client
	}
	return p._client, nil
}

// Close cleans up resources associated with the Provider
func (p *Provider) Close() error {
	if s := p._client; s != nil {
		return s.Close()
	}
	return nil
}

// Configure returns static information about the integration
func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	domain, err := adclient.GetADDomain(client)
	if err != nil {
		return nil, err
	}

	name, _ := lo.Coalesce(
		domain.DNSRoot,
		domain.Forest,
		domain.NetBIOSName,
		domain.Name,
		"Microsoft Active Directory")

	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                    name,
			CanGetTemporaryPassword: lo.ToPtr(true),
			CanUnlock:               lo.ToPtr(true),
			CanUpdateAccountsList:   lo.ToPtr(true),
		},
	}, nil
}
