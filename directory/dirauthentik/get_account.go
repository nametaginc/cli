// Copyright 2026 Nametag Inc.
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

package dirauthentik

import (
	"context"
	"net/url"

	"github.com/nametaginc/cli/diragentapi"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
	queries := []queryRequest{}

	if req.Ref.ImmutableID != nil {
		query := baseUserQuery(true)
		query.Set("uuid", *req.Ref.ImmutableID)
		queries = append(queries, queryRequest{query: query, allowPKLookup: isNumeric(*req.Ref.ImmutableID)})
	}
	if req.Ref.ID != nil {
		query := baseUserQuery(true)
		query.Set("email", *req.Ref.ID)
		queries = append(queries, queryRequest{query: query})

		query = baseUserQuery(true)
		query.Set("username", *req.Ref.ID)
		queries = append(queries, queryRequest{query: query})
	}

	if len(queries) == 0 {
		return &diragentapi.DirAgentGetAccountResponse{}, nil
	}

	accounts := []diragentapi.DirAgentAccount{}
	seen := map[string]struct{}{}

	for _, request := range queries {
		users, err := p.fetchUsers(ctx, request.query)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			account := p.accountFromUser(user)
			if account.ImmutableID == "" {
				continue
			}
			if _, ok := seen[account.ImmutableID]; ok {
				continue
			}
			seen[account.ImmutableID] = struct{}{}
			accounts = append(accounts, account)
		}

		if request.allowPKLookup && len(users) == 0 && req.Ref.ImmutableID != nil {
			user, err := p.fetchUserByPK(ctx, *req.Ref.ImmutableID)
			if err != nil {
				return nil, err
			}
			account := p.accountFromUser(*user)
			if account.ImmutableID != "" {
				if _, ok := seen[account.ImmutableID]; !ok {
					seen[account.ImmutableID] = struct{}{}
					accounts = append(accounts, account)
				}
			}
		}
	}

	return &diragentapi.DirAgentGetAccountResponse{Accounts: accounts}, nil
}

type queryRequest struct {
	query         url.Values
	allowPKLookup bool
}

func (p *Provider) accountFromUser(user apiUser) diragentapi.DirAgentAccount {
	groups := make([]diragentapi.DirAgentGroup, 0, len(user.GroupsObj))
	for _, group := range user.GroupsObj {
		if group.PK == "" && group.Name == "" {
			continue
		}
		immutableID := group.PK
		if immutableID == "" {
			immutableID = group.Name
		}
		groups = append(groups, diragentapi.DirAgentGroup{
			ImmutableID: immutableID,
			Name:        group.Name,
			Kind:        "group",
		})
	}

	account := diragentapi.DirAgentAccount{
		ImmutableID: userImmutableID(user),
		IDs:         userExternalIDs(user),
		Name:        userDisplayName(user),
		UpdatedAt:   parseAPITime(user.LastUpdated),
	}
	account.Groups = &groups
	return account
}
