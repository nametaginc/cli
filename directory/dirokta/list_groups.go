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

	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// ListGroups returns the directory groups that match the given name prefix.
func (p *Provider) ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error) {
	ctx, client, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	rv := diragentapi.DirAgentListGroupsResponse{}

	// https://developer.okta.com/docs/reference/api/groups/#list-groups
	paramOptions := []query.ParamOptions{
		query.WithSortBy("profile.name"),
	}
	if req.MaxCount != nil {
		paramOptions = append(paramOptions, query.WithLimit(*req.MaxCount))
	}
	if req.NamePrefix != nil {
		paramOptions = append(paramOptions, query.WithSearch(fmt.Sprintf("profile.name sw %q", *req.NamePrefix)))
	}
	if req.Cursor != nil {
		paramOptions = append(paramOptions, query.WithAfter(*req.Cursor))
	}

	groups, resp, err := client.Group.ListGroups(ctx, query.NewQueryParams(paramOptions...))
	if err != nil {
		return nil, fmt.Errorf("okta: failed to list groups: %w", err)
	}
	for _, group := range groups {
		rv.Groups = append(rv.Groups, diragentapi.DirAgentGroup{
			ImmutableID: group.Id,
			Kind:        group.Type,
			Name:        group.Profile.Name,
		})
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
