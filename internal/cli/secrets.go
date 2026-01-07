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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

func newSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Commands for managing secrets",
	}
	cmd.AddCommand(newSecretsEncryptCmd())
	return cmd
}

func newSecretsEncryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			envID, err := cmd.Flags().GetString("env")
			if err != nil || envID == "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "the --env flag is required\n")
				return fmt.Errorf("the --env flag is required")
			}

			client, err := NewAPIClient(cmd)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot initialize API client: %s\n", err)
				return err
			}

			plaintext, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("cannot read input: %w", err)
			}

			req := api.EncryptSecretRequest{
				Plaintext: string(plaintext),
			}

			resp, err := client.EncryptSecretWithResponse(cmd.Context(), envID,
				req)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot encrypt secret: %s\n", err)
				return err
			}
			if resp.StatusCode() != 200 {
				fmt.Fprintf(cmd.ErrOrStderr(), "cannot encrypt secret: %s\n", resp.Status())
				return fmt.Errorf("%s", resp.Status())
			}
			fmt.Fprintln(cmd.OutOrStdout(), resp.JSON200.Ciphertext)
			return nil
		},
	}
	cmd.Flags().StringP("env", "e", "", "The `environment` identifier of the self-service site.")
	return cmd
}
