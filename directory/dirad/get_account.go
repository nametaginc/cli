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
	"fmt"
	"log"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

// GetAccount fetches accounts given one of its external IDs.
//
// Because multiple accounts could match an external ID, it is possible than multiple
// accounts could be returned. The caller must handle this case, which is probably an
// error.
func (p *Provider) GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error) {
	svc, err := p.client()
	if err != nil {
		return nil, err
	}

	users, err := adclient.GetADUser(svc, adclient.GetADUserArgs{
		Identity: *req.Ref.ImmutableID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get user from ad: %w", err)
	}

	var accounts []diragentapi.DirAgentAccount
	for _, user := range *users {
		userGroups, err := adclient.GetADUserGroups(svc, user)
		if err != nil {
			return nil, fmt.Errorf("could not get groups for user %s from ad: %w", user.ObjectGUID, err)
		}

		dirGroups := lo.Map(*userGroups, func(item adclient.UserGroup, _ int) diragentapi.DirAgentGroup {
			return diragentapi.DirAgentGroup{
				ImmutableID: item.ObjectGUID,
				Name:        item.Name,
				Kind:        "group",
			}
		})

		account := diragentapi.DirAgentAccount{
			ImmutableID: user.ObjectGUID,
			IDs:         p.externalIDs(user),
			Name:        p.displayName(user),
			Groups:      &dirGroups,
		}

		log.Printf("user account: %+v", account)
		accounts = append(accounts, account)
	}

	return &diragentapi.DirAgentGetAccountResponse{Accounts: accounts}, nil
}
