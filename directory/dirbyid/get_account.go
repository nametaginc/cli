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
	"fmt"
	"log"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
	log.Printf("get_account called")

	log.Printf("req.ImmutableID: %+v", *req.Ref.ImmutableID)
	log.Printf("req.ID: %+v", *req.Ref.ID)

	// Fetch the identities that match the request.
	var identities []*byidclient.Identity
	// If the immutable ID is provided, use it to get the identity.
	if req.Ref.ImmutableID != nil {
		identity, err := p.client.GetIdentity(ctx, *req.Ref.ImmutableID)
		if err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	} else {
		// If the ref ID is provided, use it to get all identities that match.
		// This *should* be a single identity, since we're seeding the directory with a username, a unique value.
		var query string
		if p.Version == "v0" {
			query = fmt.Sprintf("username eq %q", *req.Ref.ID)
		} else {
			query = fmt.Sprintf("traits.username eq %q", *req.Ref.ID)
		}

		var pageToken *string
		for {
			resp, err := p.client.ListIdentities(ctx, &query, pageToken)
			if err != nil {
				return nil, err
			}

			identities = append(identities, resp.Identities...)

			if resp.NextPageToken == nil {
				break
			}
			pageToken = resp.NextPageToken
		}
	}

	// Fetch the groups for each identity.
	var accounts []diragentapi.DirAgentAccount
	for _, identity := range identities {
		account := toDirAgentAccount(*identity)

		var pageToken *string
		for {
			groupsResponse, err := p.client.ListIdentityGroups(ctx, identity.ID, pageToken)
			if err != nil {
				return nil, err
			}

			groups := make([]diragentapi.DirAgentGroup, 0, len(groupsResponse.Groups))
			for _, group := range groupsResponse.Groups {
				groups = append(groups, toDirAgentGroup(*group))
			}

			account.Groups = &groups

			accounts = append(accounts, account)

			// If there are no more groups, break.
			if groupsResponse.NextPageToken == nil {
				break
			}

			// If there are more groups, fetch the next page.
			pageToken = groupsResponse.NextPageToken
		}
	}

	return &diragentapi.DirAgentGetAccountResponse{
		Accounts: accounts,
	}, nil
}
