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

package dirldap

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// ListGroups returns the directory groups that match the given name prefix.
func (p *Provider) ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, fmt.Errorf("could not get client from provider: %w", err)
	}

	var namePrefix string
	if req.NamePrefix != nil {
		namePrefix = *req.NamePrefix
	}

	pagingControl := ldap.NewControlPaging(p.Config.PageSize)

	// If we have a pagination token from previous request, decode it
	if req.Cursor != nil {
		cookie, err := base64.StdEncoding.DecodeString(*req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		// Set the cookie in the paging control
		pagingControl.SetCookie(cookie)
	}

	groupSearchRequest := ldap.NewSearchRequest(
		p.Config.BaseDN, // Use the base DN instead of the prefix directly
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=group)(dn=%s*))", namePrefix), // Search for groups where CN starts with prefix
		[]string{"entryUUID", "dn"},
		[]ldap.Control{pagingControl},
	)

	groupResult, err := client.Search(groupSearchRequest)
	if err != nil {
		return nil, fmt.Errorf("error fetching group %s: %w", *req.NamePrefix, err)
	}

	userGroups := []diragentapi.DirAgentGroup{}
	if len(groupResult.Entries) > 0 {
		for _, groupEntry := range groupResult.Entries {
			userGroups = append(userGroups, diragentapi.DirAgentGroup{
				ImmutableID: groupEntry.GetAttributeValue("entryUUID"),
				Name:        groupEntry.GetAttributeValue("dn"),
				Kind:        "group",
			})
		}
	}

	rv := diragentapi.DirAgentListGroupsResponse{Groups: userGroups}
	// Get the paging response to set up the next page
	for _, control := range groupResult.Controls {
		if pagingResult, ok := control.(*ldap.ControlPaging); ok {
			if len(pagingResult.Cookie) > 0 {
				rv.NextCursor = lo.ToPtr(base64.StdEncoding.EncodeToString(pagingResult.Cookie))
			}
			break
		}
	}

	return &rv, nil
}
