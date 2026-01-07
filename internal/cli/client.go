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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
	"github.com/nametaginc/cli/internal/config"
)

// NewAPIClient returns a new API client configured as appropriate given the
// environment variables and command line flags.
func NewAPIClient(cmd *cobra.Command) (*api.ClientWithResponses, error) {
	authToken, err := getAuthToken(cmd)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClientWithResponses(getServer(cmd), api.WithRequestEditorFn(
		func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer "+authToken)
			return nil
		}),
		api.WithHTTPClient(HTTPClient),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
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
		cliConfig, err := config.ReadConfig(cmd)
		if err != nil {
			return "", err
		}
		authToken = cliConfig.Token
	}

	if authToken != "" {
		var claims jwt.RegisteredClaims
		_, _, err := new(jwt.Parser).ParseUnverified(authToken, &claims)
		if err == nil && time.Now().After(lo.FromPtr(claims.ExpiresAt).Time) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Your authentication token has expired. Do you need to run `nametag auth login` again?\n")
			os.Exit(1)
		}
	}

	if authToken == "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Cannot find an authentation token. Do you need to run `nametag auth login`?\n")
		os.Exit(1)
	}

	return authToken, nil
}
