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

// Package directory contains type definitions for directory providers.
package directory

import (
	"context"
	"fmt"

	"github.com/nametaginc/cli/diragentapi"
)

// Provider is an interface that represents a directory provider (e.g. Azure AD, Okta, etc.)
type Provider interface {
	// Configure returns static information about the integration
	Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error)

	// ListAccounts returns a partial list of accounts. Callers should use Cursor to page
	// through multiple pages of results.
	ListAccounts(ctx context.Context, req diragentapi.DirAgentListAccountsRequest) (*diragentapi.DirAgentListAccountsResponse, error)

	// GetAccount fetches accounts given one of its external IDs.
	// Because multiple accounts could match an external ID, it is possible than multiple
	// accounts could be returned. The caller must handle this case, which is probably an
	// error.
	GetAccount(ctx context.Context, req diragentapi.DirAgentGetAccountRequest) (*diragentapi.DirAgentGetAccountResponse, error)

	// ListGroups returns the directory groups that match the given name prefix.
	ListGroups(ctx context.Context, req diragentapi.DirAgentListGroupsRequest) (*diragentapi.DirAgentListGroupsResponse, error)

	// PerformOperation executes a recovery operation, unless DryRun is specified, in which case it
	// checked the preconditions for the operation.
	PerformOperation(ctx context.Context, req diragentapi.DirAgentPerformOperationRequest) (*diragentapi.DirAgentPerformOperationResponse, error)
}

// CodedError is an alias for diragentapi.DirAgentErrorResponse that
// implements the error interface so it can be returned from Provider
// methods. When this error is returned, the diragentapi.DirAgentResponse
// will have the corresponding Error field set.
type CodedError diragentapi.DirAgentErrorResponse

func (c CodedError) Error() string {
	return fmt.Sprintf("%s %s", c.Code, c.Message)
}
