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
	"github.com/nametaginc/cli/directory/dirad/adclient"
)

// performOperationGetTemporaryPassword will generate and set a temporary password required to be changed on first login.
func (p *Provider) performOperationGetTemporaryPassword(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}
	if lo.FromPtr(req.DryRun) {
		return &diragentapi.DirAgentPerformOperationResponse{}, nil
	}

	tempPassword, err := adclient.AssignTemporaryPassword(client, adclient.PasswordArgs{UserImmutableID: req.AccountImmutableID})
	if err != nil {
		return nil, err
	}

	response := &diragentapi.DirAgentPerformOperationResponse{
		TemporaryPassword: tempPassword,
	}
	return response, nil
}
