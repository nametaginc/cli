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

	"github.com/nametaginc/cli/diragentapi"
	"github.com/samber/lo"
)

type Provider struct {
	URL          string
	ClientID     string
	ClientSecret string
}

func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                  "Beyond Identity",
			CanGetPasswordLink:    lo.ToPtr(true),
			CanRemoveAllMFA:       lo.ToPtr(true),
			CanUnlock:             lo.ToPtr(true),
			CanUpdateAccountsList: lo.ToPtr(true),
		},
		ImmutableID: fmt.Sprintf("urn:agent:%s", p.URL),
	}, nil
}
