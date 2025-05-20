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

//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
)

func main() {
	requests := json.NewDecoder(os.Stdin)
	responses := json.NewEncoder(os.Stdout)

	for {
		req := diragentapi.DirAgentRequest{}
		if err := requests.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("ERROR: cannot decode request: %s", err)
		}

		if err := json.NewEncoder(os.Stderr).Encode(&req); err != nil {
			log.Fatalf("ERROR: cannot write to stderr: %s", err)
		}

		resp := diragentapi.DirAgentResponse{}
		switch {
		case req.Ping != nil:
			// ok
		case req.Configure != nil:
			resp.Configure = &diragentapi.DirAgentConfigureResponse{
				Traits: diragentapi.DirAgentTraits{
					CanGetMFABypassCode:   lo.ToPtr(true),
					CanGetPasswordLink:    lo.ToPtr(true),
					CanRemoveAllMFA:       lo.ToPtr(true),
					CanUnlock:             lo.ToPtr(true),
					CanUpdateAccountsList: lo.ToPtr(true),
					Name:                  "fake",
				},
			}
		default:
			resp.Error = &diragentapi.DirAgentErrorResponse{
				Code:    diragentapi.ConfigurationError,
				Message: "unsupported operation",
			}
		}

		if err := responses.Encode(&resp); err != nil {
			log.Fatalf("ERROR: cannot decode request: %s", err)
		}
	}
}
