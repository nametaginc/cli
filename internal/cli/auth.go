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
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

// Config represents the format of the configuration file that
// contains the authentication token and other settings.
type Config struct {
	Version string `yaml:"version"`
	Server  string `yaml:",omitempty"`
	Token   string `yaml:"token"`
}

func configPath(cmd *cobra.Command) (string, error) {
	param, err := cmd.Flags().GetString("config")
	if err != nil {
		return "", err
	}
	if param != "" {
		return param, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".config", "nametag", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return "", err
	}
	return configPath, nil
}

var (
	cachedConfig *Config
)

func readConfig(cmd *cobra.Command) (*Config, error) {
	if cachedConfig != nil {
		return cachedConfig, nil
	}
	path, err := configPath(cmd)
	if err != nil {
		return nil, err
	}
	configBuf, err := os.ReadFile(path) //nolint:gosec  // no file inclusion vulnerability; this is a client side application where it's okay to specify a path
	if os.IsNotExist(err) {
		return &Config{}, nil
	} else if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(configBuf, &config); err != nil {
		return nil, fmt.Errorf("configuration file is not valid: %s: %w", path, err)
	}
	if config.Version != "1" {
		return nil, fmt.Errorf("unsupported configuration file version: %s", config.Version)
	}

	cachedConfig = &config
	return &config, nil
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
		config, err := readConfig(cmd)
		if err != nil {
			return "", err
		}
		authToken = config.Token
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
