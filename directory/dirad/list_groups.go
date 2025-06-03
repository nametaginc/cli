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

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

// ListGroups returns the directory groups that match the given name prefix.
func (p *Provider) ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	rv := diragentapi.DirAgentListGroupsResponse{}
	groups, nextCursor, err := adclient.GetADGroups(client, adclient.GetADGroupArgs{
		NamePrefix: req.NamePrefix,
		Cursor:     req.Cursor,
	})
	if err != nil {
		return nil, err
	}

	for _, group := range *groups {
		rv.Groups = append(rv.Groups, diragentapi.DirAgentGroup{
			ImmutableID: group.ObjectGUID,
			Name:        group.Name,
		})
	}

	rv.NextCursor = nextCursor
	return &rv, nil
}
