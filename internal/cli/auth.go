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

package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

func init() {
	Root.PersistentFlags().StringP("auth-token", "t", "", "Nametag API authentication token")
}

func getAuthToken(cmd *cobra.Command) (authToken string, err error) {
	authToken, err = cmd.Flags().GetString("auth-token")
	if err != nil {
		return "", err
	}
	if authToken == "" {
		authToken = os.Getenv("NAMETAG_AUTH_TOKEN")
	}
	if authToken == "" {
		return "", fmt.Errorf("specify an authentication token with --auth-token or $NAMETAG_AUTH_TOKEN")
	}

	return authToken, nil
}

// NewAPIClient returns a new API client configured as appropriate given the
// environment variables and command line flags.
func NewAPIClient(cmd *cobra.Command) (*api.ClientWithResponses, error) {
	authToken, err := getAuthToken(cmd)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClientWithResponses(Server, api.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer "+authToken)
			return nil
		}))
	if err != nil {
		return nil, err
	}

	return client, nil
}
