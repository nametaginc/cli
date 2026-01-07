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
	"fmt"
	"log"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/nametaginc/cli/diragentapi"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
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
		"memberOf",        // Indicates group membership
	}

	// Search filter for users
	searchRequest := ldap.NewSearchRequest(
		p.Config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(entryUUID=%s)", *req.Ref.ImmutableID), // Filter for user objects
		attributes,
		nil,
	)

	result, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error fetching user %s: %w", *req.Ref.ImmutableID, err)
	}

	var accounts []diragentapi.DirAgentAccount
	for _, entry := range result.Entries {
		externalIDs := []string{}
		if entry.GetAttributeValue("mail") != "" {
			externalIDs = append(externalIDs, entry.GetAttributeValue("mail"))
		}
		if entry.GetAttributeValue("uid") != "" {
			externalIDs = append(externalIDs, entry.GetAttributeValue("uid"))
		}
		modifyTime, _ := time.Parse(timeFormat, entry.GetAttributeValue("modifyTimestamp"))

		groups := entry.GetAttributeValues("memberOf")
		userGroups := []diragentapi.DirAgentGroup{}
		for _, groupDN := range groups {
			// Search for the group to get its entryUUID
			groupSearchRequest := ldap.NewSearchRequest(
				groupDN, // Search directly using the group's DN
				ldap.ScopeBaseObject,
				ldap.NeverDerefAliases,
				0,
				0,
				false,
				"(objectClass=*)",
				[]string{"entryUUID"}, // We only need the entryUUID
				nil,
			)

			groupResult, err := client.Search(groupSearchRequest)
			if err != nil {
				return nil, fmt.Errorf("error fetching group %s: %w", groupDN, err)
			}

			if len(groupResult.Entries) > 0 {
				groupEntry := groupResult.Entries[0]
				userGroups = append(userGroups, diragentapi.DirAgentGroup{
					ImmutableID: groupEntry.GetAttributeValue("entryUUID"),
					Name:        groupDN,
					Kind:        "group",
				})
			}
		}

		account := diragentapi.DirAgentAccount{
			ImmutableID: entry.GetAttributeValue("entryUUID"),
			IDs:         externalIDs,
			Name:        entry.GetAttributeValue("cn"),
			UpdatedAt:   &modifyTime,
			Groups:      &userGroups,
		}

		log.Printf("user account: %+v", account)
		accounts = append(accounts, account)
	}

	return &diragentapi.DirAgentGetAccountResponse{Accounts: accounts}, nil
}
