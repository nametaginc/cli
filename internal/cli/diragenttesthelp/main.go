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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

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
					CanGetMFALink:         lo.ToPtr(true),
					CanGetPasswordLink:    lo.ToPtr(true),
					CanRemoveAllMFA:       lo.ToPtr(true),
					CanUnlock:             lo.ToPtr(true),
					CanUpdateAccountsList: lo.ToPtr(true),
					Name:                  "fake",
				},
			}
		case req.ListAccounts != nil:
			const totalFakeAccounts = 500
			const testPageSize = 250 // matches client.go
			const groupsPerAccount = 5

			offset := 0
			if req.ListAccounts.Cursor != nil {
				parsed, err := strconv.Atoi(*req.ListAccounts.Cursor)
				if err != nil {
					log.Fatalf("ERROR: bad cursor: %v", err)
				}
				offset = parsed
			}

			end := offset + testPageSize
			if end > totalFakeAccounts {
				end = totalFakeAccounts
			}

			var accounts []diragentapi.DirAgentAccount
			for i := offset; i < end; i++ {
				groups := make([]diragentapi.DirAgentGroup, groupsPerAccount)
				for g := range groups {
					groups[g] = diragentapi.DirAgentGroup{
						ImmutableID: fmt.Sprintf("f00dcafe-%04d-4000-8000-%012d", g, i),
						Name:        fmt.Sprintf("Personal Access Group %d for user %d", g, i),
						Kind:        "security group",
					}
				}

				accounts = append(accounts, diragentapi.DirAgentAccount{
					ImmutableID: fmt.Sprintf("d3adbeef-0000-4000-8000-%012d", i),
					IDs: []string{
						fmt.Sprintf("jdoe%d@example.com", i),
						fmt.Sprintf("jdoe%d", i),
					},
					Name:      fmt.Sprintf("Jonathan Middleton Doe %d", i),
					UpdatedAt: lo.ToPtr(time.Now()),
					Groups:    lo.ToPtr(groups),
				})
			}

			resp.ListAccounts = &diragentapi.DirAgentListAccountsResponse{
				Accounts: accounts,
			}
			if end < totalFakeAccounts {
				resp.ListAccounts.NextCursor = lo.ToPtr(strconv.Itoa(end))
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
