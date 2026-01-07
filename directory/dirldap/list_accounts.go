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
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

// ListAccounts returns a partial list of accounts. Callers should use Cursor to page
// through multiple pages of results.
func (p *Provider) ListAccounts(ctx context.Context, req diragentapi.DirAgentListAccountsRequest) (*diragentapi.DirAgentListAccountsResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, fmt.Errorf("could not get client from provider: %w", err)
	}

	attributes := []string{
		"entryUUID",       // Immutable ID
		"uid",             // Username
		"mail",            // Email address
		"cn",              // Common Name
		"modifyTimestamp", // When last modified
	}

	var filterString string
	if req.UpdatedAfter != nil {
		timeFilter := req.UpdatedAfter.Format(timeFormat)
		filterString = fmt.Sprintf("(&(objectClass=inetOrgPerson)(modifyTimestamp>=%s))", timeFilter)
	} else {
		filterString = "(objectClass=inetOrgPerson)"
	}

	// Create paging control
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

	// Search filter for users
	searchRequest := ldap.NewSearchRequest(
		p.Config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filterString, // Filter for user objects
		attributes,
		[]ldap.Control{pagingControl},
	)

	result, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %w", err)
	}

	rv := diragentapi.DirAgentListAccountsResponse{}
	for _, entry := range result.Entries {
		externalIDs := []string{}
		if entry.GetAttributeValue("mail") != "" {
			externalIDs = append(externalIDs, entry.GetAttributeValue("mail"))
		}
		if entry.GetAttributeValue("uid") != "" {
			externalIDs = append(externalIDs, entry.GetAttributeValue("uid"))
		}
		modifyTime, _ := time.Parse(timeFormat, entry.GetAttributeValue("modifyTimestamp"))
		account := diragentapi.DirAgentAccount{
			ImmutableID: entry.GetAttributeValue("entryUUID"),
			IDs:         externalIDs,
			Name:        entry.GetAttributeValue("cn"),
			UpdatedAt:   &modifyTime,
		}

		rv.Accounts = append(rv.Accounts, account)
	}

	// Get the paging response to set up the next page
	for _, control := range result.Controls {
		if pagingResult, ok := control.(*ldap.ControlPaging); ok {
			if len(pagingResult.Cookie) > 0 {
				rv.NextCursor = lo.ToPtr(base64.StdEncoding.EncodeToString(pagingResult.Cookie))
			}
			break
		}
	}

	return &rv, nil
}
