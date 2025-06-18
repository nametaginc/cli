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

package dirbyid

import (
	"context"
	"log"

	"github.com/nametaginc/cli/diragentapi"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
	log.Printf("get_account called")
	return &diragentapi.DirAgentGetAccountResponse{
		Accounts: []diragentapi.DirAgentAccount{
			{
				ImmutableID: "123",
				IDs:         []string{"123"},
				Name:        "John Doe",
			},
		},
	}, nil
}
