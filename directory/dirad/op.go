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
	"fmt"

	"github.com/nametaginc/cli/diragentapi"
)

// PerformOperation performs the specified recovery operation
func (p *Provider) PerformOperation(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	switch req.Operation {
	case diragentapi.GetTemporaryPassword:
		return p.performOperationGetTemporaryPassword(ctx, req)
	case diragentapi.Unlock:
		return p.performOperationUnlock(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported operation %s", req.Operation)
	}
}
