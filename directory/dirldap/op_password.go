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
	"strconv"

	"github.com/go-ldap/ldap/v3"
	"github.com/samber/lo"
	"github.com/sethvargo/go-password/password"

	"github.com/nametaginc/cli/diragentapi"
)

const defaultPasswordLength = 6

// performOperationGetTemporaryPassword will generate and set a temporary password required to be changed on first login.
func (p *Provider) performOperationGetTemporaryPassword(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, fmt.Errorf("could not get client from provider: %w", err)
	}

	if lo.FromPtr(req.DryRun) {
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	searchRequest := ldap.NewSearchRequest(
		p.Config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(entryUUID=%s)", req.AccountImmutableID), // Filter for user objects
		[]string{
			"pwdPolicySubentry", // Points to the specific password policy entry
		},
		nil,
	)

	dnResult, err := client.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if dnResult.Entries != nil && len(dnResult.Entries) != 1 {
		return nil, fmt.Errorf("expected exactly one result, got %d", len(dnResult.Entries))
	}

	entry := dnResult.Entries[0]
	var pwdMinLength int
	// First, get the password policy DN from the user entry
	policyDN := entry.GetAttributeValue("pwdPolicySubentry")
	// If the user does not have a pwd policy associated or a default configured,
	// we generate a password based on a fixed length.
	if policyDN == "" && p.Config.DefaultPasswordPolicyDN == "" {
		pwdMinLength = defaultPasswordLength
	} else {
		if policyDN == "" {
			policyDN = p.Config.DefaultPasswordPolicyDN
		}

		// Create a new search request for the password policy entry
		policySearchRequest := ldap.NewSearchRequest(
			policyDN,             // Use the policy DN directly
			ldap.ScopeBaseObject, // We want just this entry
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			"(objectClass=*)", // Get the policy entry
			[]string{
				"pwdMinLength",
				"pwdAllowUserChange",
				"pwdMaxFailure",           // Max login attempts
				"pwdLockoutDuration",      // How long account stays locked
				"pwdFailureCountInterval", // Time window for counting failures
			},
			nil,
		)

		policyResult, err := client.Search(policySearchRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch password policy: %w", err)
		}

		if len(policyResult.Entries) != 1 {
			return nil, fmt.Errorf("expected exactly one policy entry, got %d", len(policyResult.Entries))
		}

		policyEntry := policyResult.Entries[0]
		pwdMinLengthStr := policyEntry.GetAttributeValue("pwdMinLength")
		if pwdMinLengthStr == "" {
			return nil, fmt.Errorf("failed to fetch password policy: missing pwdMinLength attribute")
		}

		pwdMinLength, err = strconv.Atoi(pwdMinLengthStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pwdMinLength: %w", err)
		}
	}

	// Create modify request for password change
	passwordModify := ldap.NewModifyRequest(entry.DN, nil)

	tempPassword, err := password.Generate(pwdMinLength, 1, 1, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Hash the password using SSHA (Salted SHA1) - recommended for OpenLDAP
	passwordModify.Replace("userPassword", []string{tempPassword})

	// Set pwdReset to TRUE to force password change on next login
	passwordModify.Replace("pwdReset", []string{"TRUE"})

	// Create and add the password policy control
	pPolicyControl := ldap.NewControlString(
		"1.3.6.1.4.1.42.2.27.8.5.1", // Password Policy Control OID
		true,                        // Criticality set to true
		"",                          // No control value needed
	)
	passwordModify.Controls = append(passwordModify.Controls, pPolicyControl)

	err = client.Modify(passwordModify)
	if err != nil {
		return nil, fmt.Errorf("failed to modify password and set reset flag: %w", err)
	}

	response := &diragentapi.DirAgentPerformOperationResponse{
		TemporaryPassword: &tempPassword,
	}
	return response, nil
}
