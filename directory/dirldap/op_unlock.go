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

	"github.com/go-ldap/ldap/v3"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
)

// performOperationUnlock unlock a user account.
func (p *Provider) performOperationUnlock(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, fmt.Errorf("could not get client from provider: %w", err)
	}

	unlockAttribute := "pwdAccountLockedTime"
	searchRequest := ldap.NewSearchRequest(
		p.Config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(entryUUID=%s)", req.AccountImmutableID), // Filter for user objects
		[]string{
			unlockAttribute,
		},
		nil,
	)

	result, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}

	if result.Entries != nil && len(result.Entries) != 1 {
		return nil, fmt.Errorf("expected exactly one result, got %d", len(result.Entries))
	}

	userEntry := result.Entries[0]
	accountLockTime := userEntry.GetAttributeValue(unlockAttribute)
	if accountLockTime == "" {
		return nil, directory.CodedError{
			Code:    diragentapi.UnsupportedAccountState,
			Message: "account is not locked",
		}
	}

	if lo.FromPtr(req.DryRun) {
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	modify := ldap.NewModifyRequest(userEntry.DN, nil)

	// Delete the pwdAccountLockedTime attribute
	modify.Delete(unlockAttribute, []string{})

	// Add ppolicy control to get additional information
	pPolicyControl := ldap.NewControlString(
		"1.3.6.1.4.1.42.2.27.8.5.1", // Password Policy Control OID
		true,                        // Criticality
		"",                          // No value needed
	)
	modify.Controls = append(modify.Controls, pPolicyControl)

	err = client.Modify(modify)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock account: %w", err)
	}

	return &diragentapi.DirAgentPerformOperationResponse{}, nil
}
