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
	"errors"
	"fmt"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"

	"github.com/nametaginc/cli/diragentapi"
)

// PerformOperation performs the specified recovery operation
func (p *Provider) PerformOperation(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error) {
	switch req.Operation {
	case diragentapi.GetPasswordLink:
		return p.performOperationGetPasswordLink(ctx, req)
	case diragentapi.RemoveAllMFA:
		return p.performOperationRemoveAllMfa(ctx, req)
	case diragentapi.Unlock:
		return p.performOperationUnlock(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported operation %s", req.Operation)
	}
}

func (p *Provider) filterAPIError(err error) error {
	var oktaError *okta.Error
	if errors.As(err, &oktaError) {
		if oktaError.ErrorSummary != "" {
			errStr := []string{oktaError.ErrorSummary}
			for _, errorCause := range oktaError.ErrorCauses {
				if es, ok := errorCause["errorSummary"]; ok {
					if es := es.(string); ok {
						errStr = append(errStr, es)
					}
				}
			}
			return fmt.Errorf("%s", strings.Join(errStr, ": "))
		}
	}
	return err
}
