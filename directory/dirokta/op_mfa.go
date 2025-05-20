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

package dirokta

import (
	"context"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
)

func (p *Provider) performOperationRemoveAllMfa(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	ctx, oktaClient, err := p.client(ctx)
	if err != nil {
		return nil, err
	}

	if lo.FromPtr(req.DryRun) {
		factors, _, err := oktaClient.UserFactor.ListFactors(ctx, req.AccountImmutableID)
		if err != nil {
			return nil, err
		}

		if len(factors) == 0 {
			return nil, directory.CodedError{
				Code:    diragentapi.UnsupportedAccountState,
				Message: "no MFA factors to remove",
			}
		}
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	if _, err := oktaClient.User.ResetFactors(ctx, req.AccountImmutableID); err != nil {
		return nil, p.filterAPIError(err)
	}

	return &diragentapi.DirAgentPerformOperationResponse{}, nil
}
