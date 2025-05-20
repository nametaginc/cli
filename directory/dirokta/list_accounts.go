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
	"net/url"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

const oktaTimeFormat = "2006-01-02T15:04:05.000Z"

// ListAccounts returns a partial list of accounts. Callers should use Cursor to page
// through multiple pages of results.
func (p *Provider) ListAccounts(ctx context.Context, req diragentapi.DirAgentListAccountsRequest) (*diragentapi.DirAgentListAccountsResponse, error) {
	ctx, client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	rv := diragentapi.DirAgentListAccountsResponse{}

	paramOptions := []query.ParamOptions{
		query.WithLimit(250),
	}
	if req.UpdatedAfter != nil {
		paramOptions = append(paramOptions, query.WithFilter(
			fmt.Sprintf("lastUpdated gt \"%s\"",
				req.UpdatedAfter.Format(oktaTimeFormat))))
	}
	if req.Cursor != nil {
		paramOptions = append(paramOptions, query.WithAfter(*req.Cursor))
	}

	users, resp, err := client.User.ListUsers(ctx, query.NewQueryParams(paramOptions...))
	if err != nil {
		return nil, fmt.Errorf("okta: failed to list users: %w", err)
	}

	for _, user := range users {
		account := diragentapi.DirAgentAccount{
			ImmutableID: user.Id,
			IDs:         p.externalIDs(user),
			Name:        p.displayName(user),
			UpdatedAt:   user.LastUpdated,
		}
		if user.Profile != nil {
			if birthDate, ok := (*user.Profile)["birthdate"]; ok {
				if birthDateStr, ok := birthDate.(string); ok {
					account.BirthDate = &birthDateStr
				}
			}
		}

		rv.Accounts = append(rv.Accounts, account)
	}

	if resp.HasNextPage() {
		nextURL, err := url.Parse(resp.NextPage)
		if err != nil {
			return nil, fmt.Errorf("expected next URL to be valid, got %q: %w", resp.NextPage, err)
		}
		nextCursor := nextURL.Query().Get("after")
		if nextCursor == "" {
			return nil, fmt.Errorf("expected next URL to have an `after` parameter, got %q", resp.NextPage)
		}

		rv.NextCursor = lo.ToPtr(nextCursor)
	}

	return &rv, nil
}

func (p *Provider) externalIDs(user *okta.User) []string {
	var rv []string
	profile := user.Profile
	if profile != nil {
		for _, fieldName := range []string{"login", "email", "secondEmail"} {
			value, _ := (*profile)[fieldName].(string)
			if value != "" {
				rv = append(rv, value)
			}
		}
	}
	return rv
}

func (p *Provider) displayName(user *okta.User) string {
	profile := user.Profile
	if profile == nil {
		return ""
	}
	firstName, _ := (*profile)["firstName"].(string)
	lastName, _ := (*profile)["lastName"].(string)
	return strings.TrimSpace(firstName + " " + lastName)
}
