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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jpillora/backoff"
	"github.com/pkg/browser"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/box"

	"github.com/nametaginc/cli/internal/config"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Commands for authenticating to Nametag",
	}
	cmd.AddCommand(newAuthLoginCmd())

	return cmd
}

var browserOpenURL = browser.OpenURL

var randReader = rand.Reader

func newAuthLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to Nametag",
		RunE: func(cmd *cobra.Command, args []string) error {
			noBrowser, err := cmd.Flags().GetBool("no-browser")
			if err != nil {
				return err
			}

			publicKey, privateKey, err := box.GenerateKey(randReader)
			if err != nil {
				return err
			}

			server := getServer(cmd)
			url := server + "/cli/login/" + base64.RawURLEncoding.EncodeToString(publicKey[:])
			fmt.Fprintf(cmd.OutOrStdout(), "If your browser does not open automatically, then navigate to the following URL to authenticate to Nametag\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", url)
			if !noBrowser {
				_ = browserOpenURL(url)
			}

			bo := backoff.Backoff{
				Min: time.Second,
				Max: 10 * time.Second,
			}
			for {
				url := server + "/api/cli/login/" + base64.RawURLEncoding.EncodeToString(publicKey[:])
				req, err := http.NewRequestWithContext(cmd.Context(), "GET", url, nil)
				if err != nil {
					return err
				}
				req.Header.Add("Accept", "application/json")
				resp, err := HTTPClient.Do(req)
				if err != nil {
					cmd.Printf("failed to fetch authentication token: %s\n", err)
					return err
				}
				if resp.StatusCode >= 400 {
					cmd.Printf("failed to fetch authentication token: %s\n", resp.Status)
					return fmt.Errorf("failed to fetch authentication token: %s", resp.Status)
				}

				if resp.StatusCode == http.StatusNoContent {
					if bo.Attempt() > 4 {
						fmt.Fprintf(cmd.OutOrStdout(), "Authentication timed out. Run this command again to retry authentication.\n")
						os.Exit(1)
					}

					time.Sleep(bo.Duration())
					continue
				}

				encryptedResp, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				plaintextToken, ok := box.OpenAnonymous(nil, encryptedResp, publicKey, privateKey)
				if !ok {
					return err
				}

				configPath, err := config.GetPath(cmd)
				if err != nil {
					return err
				}

				config := config.Config{
					Version: "1",
					Server:  lo.If(server == defaultServer, "").Else(server),
					Token:   string(plaintextToken),
				}
				configBuf, err := yaml.Marshal(config)
				if err != nil {
					return err
				}
				if err := os.WriteFile(configPath, configBuf, 0600); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "You are now logged in. You can run other nametag commands now.\n")
				return nil
			}
		},
	}
	cmd.Flags().Bool("no-browser", false, "Disable automatic browser opening")
	return cmd
}
