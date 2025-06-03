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

package dirad

import (
	"context"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

// performOperationUnlock will unlock a user account
func (p *Provider) performOperationUnlock(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	args := adclient.UnlockArgs{UserImmutableID: req.AccountImmutableID}
	accountLocked, err := adclient.IsAccountLocked(client, args)
	if err != nil {
		return nil, err
	}

	if !*accountLocked {
		return nil, directory.CodedError{
			Code:    diragentapi.UnsupportedAccountState,
			Message: "account is not locked",
		}
	}
	if lo.FromPtr(req.DryRun) {
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	if err = adclient.UnlockAccount(client, adclient.UnlockArgs{UserImmutableID: req.AccountImmutableID}); err != nil {
		return nil, err
	}

	return &diragentapi.DirAgentPerformOperationResponse{}, nil
}
