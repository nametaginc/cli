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
)

// UnlockArgs is request arguments to unlock functions
type UnlockArgs struct {
	UserImmutableID string
}

// IsAccountLocked will return a boolean indicating account lockout status
func IsAccountLocked(s Client, args UnlockArgs) (*bool, error) {
	cmdString := fmt.Sprintf("Get-ADUser -Identity %s -Properties LockedOut | Select-Object LockedOut | ConvertTo-Json", args.UserImmutableID)
	stdout, err := s.Execute(cmdString)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal([]byte(stdout), &user); err != nil {
		return nil, err
	}

	return &user.LockedOut, nil
}

// UnlockAccount will unlock a user account
func UnlockAccount(s Client, args UnlockArgs) error {
	cmdString := fmt.Sprintf("Unlock-ADAccount -Identity %s", args.UserImmutableID)
	_, err := s.Execute(cmdString)
	if err != nil {
		return err
	}

	return nil
}
