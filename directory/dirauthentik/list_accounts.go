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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nametaginc/cli/diragentapi"
)

// ListAccounts returns a partial list of accounts. Callers should use Cursor to page
// through multiple pages of results.
func (p *Provider) ListAccounts(ctx context.Context, req diragentapi.DirAgentListAccountsRequest) (*diragentapi.DirAgentListAccountsResponse, error) {
	page := 1
	if req.Cursor != nil {
		parsed, err := strconv.Atoi(*req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if parsed < 1 {
			return nil, fmt.Errorf("invalid cursor: %q", *req.Cursor)
		}
		page = parsed
	}

	query := baseUserQuery(false)
	query.Set("page", strconv.Itoa(page))
	query.Set("ordering", "last_updated")

	if req.UpdatedAfter != nil {
		query.Set("last_updated__gt", req.UpdatedAfter.UTC().Format(time.RFC3339Nano))
	}

	var resp userListResponse
	if err := p.doJSON(ctx, http.MethodGet, "core/users/", query, nil, &resp); err != nil {
		return nil, err
	}

	accounts := make([]diragentapi.DirAgentAccount, 0, len(resp.Results))
	for _, user := range resp.Results {
		immutableID := userImmutableID(user)
		if immutableID == "" {
			continue
		}
		updatedAt := parseAPITime(user.LastUpdated)
		account := diragentapi.DirAgentAccount{
			ImmutableID: immutableID,
			IDs:         userExternalIDs(user),
			Name:        userDisplayName(user),
			UpdatedAt:   updatedAt,
		}
		accounts = append(accounts, account)
	}

	response := diragentapi.DirAgentListAccountsResponse{Accounts: accounts}
	if resp.Pagination.Next != nil && *resp.Pagination.Next > 0 {
		next := strconv.Itoa(*resp.Pagination.Next)
		response.NextCursor = &next
	}

	return &response, nil
}
