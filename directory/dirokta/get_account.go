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

package dirokta

import (
	"context"
	"fmt"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
	ctx, client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	queryExprs := []string{}
	if req.Ref.ImmutableID != nil {
		queryExprs = append(queryExprs, fmt.Sprintf("(id eq %q)", *req.Ref.ImmutableID))
	}
	if req.Ref.ID != nil {
		queryExprs = append(queryExprs, fmt.Sprintf("(profile.login eq %q)", *req.Ref.ID))
		queryExprs = append(queryExprs, fmt.Sprintf("(profile.email eq %q)", *req.Ref.ID))
		queryExprs = append(queryExprs, fmt.Sprintf("(profile.secondEmail eq %q)", *req.Ref.ID))
	}

	queryExpr := strings.Join(queryExprs, " or ")

	var users []*okta.User
	usersPage, resp, err := client.User.ListUsers(ctx,
		query.NewQueryParams(
			query.WithLimit(250),
			query.WithSearch(queryExpr),
		))
	if err != nil {
		return nil, fmt.Errorf("okta: failed to list users: %w", err)
	}
	users = append(users, usersPage...)
	for resp.HasNextPage() {
		var usersPage []*okta.User
		resp, err = resp.Next(ctx, &users)
		if err != nil {
			return nil, fmt.Errorf("okta: failed to list users: %w", err)
		}
		users = append(users, usersPage...)
	}

	var accounts []diragentapi.DirAgentAccount

	for _, user := range users {
		groups, _, err := client.User.ListUserGroups(ctx, user.Id)
		if err != nil {
			return nil, fmt.Errorf("okta: failed to list groups for user: %w", err)
		}

		userGroups := lo.Map(groups, func(item *okta.Group, _ int) diragentapi.DirAgentGroup {
			return diragentapi.DirAgentGroup{
				ImmutableID: item.Id,
				Name:        lo.FromPtr(item.Profile).Name,
				Kind:        "group",
			}
		})
		accounts = append(accounts, diragentapi.DirAgentAccount{
			ImmutableID: user.Id,
			IDs:         p.externalIDs(user),
			Name:        p.displayName(user),
			Groups:      &userGroups,
		})
	}
	return &diragentapi.DirAgentGetAccountResponse{Accounts: accounts}, nil
}
