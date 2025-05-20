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

package diragent

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
)

// RunWorker implements the worker process for a directory agent. It
// handles accepting and processing requests from the server. It returns
// when ctx is canceled.
func RunWorker(ctx context.Context, provider directory.Provider) error {
	input := json.NewDecoder(os.Stdin)
	output := json.NewEncoder(os.Stdout)

	for {
		req := diragentapi.DirAgentRequest{}
		if err := input.Decode(&req); err != nil {
			return err
		}
		resp := workerDoRequest(ctx, provider, req)
		if err := output.Encode(&resp); err != nil {
			return err
		}
	}
}

func workerDoRequest(ctx context.Context, provider directory.Provider, req diragentapi.DirAgentRequest) *diragentapi.DirAgentResponse {
	handleError := func(err error) *diragentapi.DirAgentResponse {
		resp := &diragentapi.DirAgentResponse{
			Error: &diragentapi.DirAgentErrorResponse{
				Code:    diragentapi.InternalError,
				Message: err.Error(),
			},
		}
		var codedErr directory.CodedError
		if errors.As(err, &codedErr) {
			resp.Error = lo.ToPtr(diragentapi.DirAgentErrorResponse(codedErr))
		}
		return resp
	}

	switch {
	case req.Ping != nil:
		return &diragentapi.DirAgentResponse{}

	case req.Configure != nil:
		resp, err := provider.Configure(ctx, *req.Configure)
		if err != nil {
			return handleError(err)
		}
		return &diragentapi.DirAgentResponse{Configure: resp}

	case req.GetAccount != nil:
		resp, err := provider.GetAccount(ctx, *req.GetAccount)
		if err != nil {
			return handleError(err)
		}
		return &diragentapi.DirAgentResponse{GetAccount: resp}

	case req.ListAccounts != nil:
		resp, err := provider.ListAccounts(ctx, *req.ListAccounts)
		if err != nil {
			return handleError(err)
		}
		return &diragentapi.DirAgentResponse{ListAccounts: resp}

	case req.ListGroups != nil:
		resp, err := provider.ListGroups(ctx, *req.ListGroups)
		if err != nil {
			return handleError(err)
		}
		return &diragentapi.DirAgentResponse{ListGroups: resp}

	case req.PerformOperation != nil:
		resp, err := provider.PerformOperation(ctx, *req.PerformOperation)
		if err != nil {
			return handleError(err)
		}
		return &diragentapi.DirAgentResponse{PerformOperation: resp}
	default:
		return &diragentapi.DirAgentResponse{
			Error: &diragentapi.DirAgentErrorResponse{
				Code:    diragentapi.InternalError,
				Message: "unknown operation",
			},
		}
	}
}
