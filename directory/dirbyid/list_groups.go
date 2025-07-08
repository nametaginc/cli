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

// ListGroups returns the directory groups that match the given name prefix.
func (p *Provider) ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error) {
	log.Printf("list_groups called")

	groupsResponse, err := p.client.ListGroups(ctx, req.Cursor)
	if err != nil {
		return nil, err
	}

	groups := make([]diragentapi.DirAgentGroup, 0, len(groupsResponse.Groups))
	for _, group := range groupsResponse.Groups {
		groups = append(groups, toDirAgentGroup(*group))
	}

	return &diragentapi.DirAgentListGroupsResponse{
		Groups:     groups,
		NextCursor: groupsResponse.NextPageToken,
	}, nil
}
