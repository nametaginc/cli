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
	"strconv"
	"strings"

	"github.com/nametaginc/cli/diragentapi"
)

// ListGroups returns the directory groups that match the given name prefix.
func (p *Provider) ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error) {
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

	query := baseGroupQuery()
	query.Set("page", strconv.Itoa(page))
	query.Set("ordering", "name")

	var prefix string
	if req.NamePrefix != nil {
		prefix = *req.NamePrefix
		query.Set("search", prefix)
	}

	resp, err := p.fetchGroups(ctx, query)
	if err != nil {
		return nil, err
	}

	groups := make([]diragentapi.DirAgentGroup, 0, len(resp.Results))
	lowerPrefix := strings.ToLower(prefix)
	for _, group := range resp.Results {
		if prefix != "" && !strings.HasPrefix(strings.ToLower(group.Name), lowerPrefix) {
			continue
		}
		immutableID := group.PK
		if immutableID == "" {
			immutableID = group.Name
		}
		if immutableID == "" {
			continue
		}
		groups = append(groups, diragentapi.DirAgentGroup{
			ImmutableID: immutableID,
			Name:        group.Name,
			Kind:        "group",
		})
	}

	response := diragentapi.DirAgentListGroupsResponse{Groups: groups}
	if resp.Pagination.Next != nil && *resp.Pagination.Next > 0 {
		next := strconv.Itoa(*resp.Pagination.Next)
		response.NextCursor = &next
	}

	return &response, nil
}
