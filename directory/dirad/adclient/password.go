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

package adclient

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sethvargo/go-password/password"
)

// PasswordPolicy is the representation of the policy in the system
type PasswordPolicy struct {
	MinPasswordLength *int `json:"MinPasswordLength,omitempty"`
}

// PasswordArgs is the args to password functions
type PasswordArgs struct {
	UserImmutableID string
}

// AssignTemporaryPassword will reset a users password with a temporary password that will be required to reset on login.
// There is no native way to have AD generate and reset with a password. It must be supplied by the admin during reset time.
// We figure out what password policy applies to the user and make a password according to the policy.
// We always operate on the assumption that `ComplexityEnabled` is true.
// This means the password has to have at least 3 out of Uppercase, Lowercase, Number, Special Characters.
func AssignTemporaryPassword(s Client, args PasswordArgs) (*string, error) {
	passwordPolicy, err := GetPasswordPolicy(s, args.UserImmutableID)
	if err != nil {
		return nil, fmt.Errorf("could not get password policy: %w", err)
	}

	res, err := password.Generate(*passwordPolicy.MinPasswordLength, 1, 1, false, false)
	if err != nil {
		return nil, err
	}

	cmdString := fmt.Sprintf("Set-ADAccountPassword -Identity %s -Reset -NewPassword (ConvertTo-SecureString '%s' -AsPlainText -Force); Set-ADUser -Identity '%s' -ChangePasswordAtLogon $true", args.UserImmutableID, res, args.UserImmutableID)
	_, err = s.Execute(cmdString)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetPasswordPolicy gets a password policy that applies to the given user.
// It checks if the user is a member of any password policy groups and returns the one with the lowest precedence.
// If there are no groups the user is a member one then the default password policy is retrieved.
func GetPasswordPolicy(s Client, immutableID string) (*PasswordPolicy, error) {
	escapedID := strings.ReplaceAll(immutableID, "'", "''")
	cmdString := fmt.Sprintf("$userDN=(Get-ADUser -Identity '%s').DistinguishedName;$groupDNs=(Get-ADUser -Identity '%s' -Properties MemberOf).MemberOf;$appliesToList=$groupDNs+$userDN;Get-ADFineGrainedPasswordPolicy -Filter * | Where-Object {$_.AppliesTo -match ($appliesToList -join '|')} | Sort-Object Precedence | Select-Object -First 1 | Select-Object MinPasswordLength | ConvertTo-Json", escapedID, escapedID)
	stdout, err := s.Execute(cmdString)
	if err != nil {
		return nil, err
	}

	passwordPolicy := PasswordPolicy{}
	if len(stdout) != 0 {
		if err = json.Unmarshal([]byte(stdout), &passwordPolicy); err != nil {
			return nil, err
		}

		return &passwordPolicy, nil
	}

	// If we don't find any Fine Grained Password Policy for the user, we use the default one.
	cmdString = "Get-ADDefaultDomainPasswordPolicy | Select-Object -Property MinPasswordLength | ConvertTo-Json"
	stdout, err = s.Execute(cmdString)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(stdout), &passwordPolicy); err != nil {
		return nil, err
	}
	return &passwordPolicy, nil
}
