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

package dirad

import (
	"context"
	"fmt"
	"time"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

const adTimeFormat = "2006-01-02T15:04:05.000Z"

// ListAccounts returns a partial list of accounts. Callers should use Cursor to page
// through multiple pages of results.
func (p *Provider) ListAccounts(ctx context.Context, req diragentapi.DirAgentListAccountsRequest) (*diragentapi.DirAgentListAccountsResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	rv := diragentapi.DirAgentListAccountsResponse{}
	users, nextCursor, err := adclient.ListADUsers(client, adclient.ListADUsersArgs{
		UpdatedAfter: req.UpdatedAfter,
		Cursor:       req.Cursor,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list AD users: %w", err)
	}

	for _, user := range *users {
		// Parse the time string into a time.Time object
		parsedTime, err := time.Parse(adTimeFormat, user.WhenChanged)
		if err != nil {
			return nil, fmt.Errorf("failed to parse update time for user %s: %w", user.ObjectGUID, err)
		}

		account := diragentapi.DirAgentAccount{
			ImmutableID: user.ObjectGUID,
			IDs:         p.externalIDs(user),
			Name:        p.displayName(user),
			UpdatedAt:   &parsedTime,
		}

		rv.Accounts = append(rv.Accounts, account)
	}

	rv.NextCursor = nextCursor

	return &rv, nil
}

func (p *Provider) externalIDs(user *adclient.User) []string {
	var rv []string
	if user.EmailAddress != "" {
		rv = append(rv, user.EmailAddress)
	}

	if user.SamAccountName != "" {
		rv = append(rv, user.SamAccountName)
	}

	return rv
}

func (p *Provider) displayName(user *adclient.User) string {
	return user.Name
}
