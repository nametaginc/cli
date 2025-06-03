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

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run a directory agent",
		Long: `Run a directory agent. 
A directory agent performs operations on behalf of Nametag such as listing accounts, 
resetting passwords, resetting MFA, and unlocking accounts. Running a directory 
agent allows you use custom directories with Nametag, shield your directory credentials 
from Nametag, or customize the behavior of already-supported directories.
To authenticate, you must specify an agent token. To get a token, use 
'nametag directory agent register'.
This command connects to Nametag and relays requests and responses between the server 
and an agent worker process that it manages.  The worker is specified by the --command flag.
Nametag comes with several built-in workers, such as Okta and Entra, which are implemented 
as subcommands of this command.
    nametag directory agent --agent-token <token> --command "nametag directory agent okta --okta-url https://example.okta.com --okta-token 1234567890"
For convenience, the built-in workers can be invoked directly. For example, the following 
command is equivalent to the one above:
	NAMETAG_AGENT_TOKEN="abcd" \
	OKTA_TOKEN="1234567890" \
	OKTA_URL="https://example.okta.com" \
    nametag directory agnet okta
An agent worker can be anything that reads JSON requests from stdin and writes JSON responses 
to stdout. The command you specify is invoked via your system shell. The environment variable 
NAMETAG_AGENT_WORKER is set to "true" when the agent is invoked as a worker process. 
For example:
    NAMETAG_AGENT_TOKEN="abcd" nametag directory agent --command "my-custom-worker"
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentToken, err := cmd.Flags().GetString("agent-token")
			if err != nil {
				return err
			}
			if agentToken == "" {
				agentToken = os.Getenv("NAMETAG_AGENT_TOKEN")
			}
			if agentToken == "" {
				return fmt.Errorf("no token provided. please specify the token using --agent-token flag or NAMETAG_AGENT_TOKEN environment variable")
			}
			command, err := cmd.Flags().GetString("command")
			if err != nil {
				return err
			}

			svc := diragent.Service{
				Server:    getServer(cmd),
				AuthToken: agentToken,
				Command:   command,
				Stderr:    cmd.ErrOrStderr(),
			}
			return svc.Run(cmd.Context())
		},
	}
	cmd.Flags().String("agent-token", "", "Nametag directory agent authentication token")
	cmd.Flags().String("command", "", "Command to run")
	_ = cmd.MarkFlagRequired("command")

	cmd.AddCommand(newDirAgentRegisterCmd())
	cmd.AddCommand(newDirAgentOktaCmd())
	cmd.AddCommand(newDirAgentADCmd())
	cmd.AddCommand(newDirAgentRegenerateTokenCmd())
	return cmd
}

func newDirAgentRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a directory agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			envID, err := cmd.Flags().GetString("env")
			if err != nil {
				return err
			}
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			resp, err := client.CreateDirectoryWithResponse(cmd.Context(),
				api.CreateDirectoryRequest{
					Kind: api.DirectoryKindCustom,
					Env:  envID,
				},
			)
			if err != nil {
				return err
			}
			if resp.StatusCode() != 200 {
				return fmt.Errorf("cannot create directory: %s", resp.Status())
			}
			logo, err := cmd.Flags().GetString("logo")
			if err != nil {
				return err
			}
			if logo != "" {
				var requestBody bytes.Buffer
				writer := multipart.NewWriter(&requestBody)
				file, err := os.Open(logo) // #nosec G304
				if err != nil {
					return fmt.Errorf("failed to open logo file: %w", err)
				}
				defer func() { _ = file.Close() }()
				part, err := writer.CreateFormFile("logo", logo)
				if err != nil {
					return fmt.Errorf("failed to create form file for image: %w", err)
				}
				_, err = io.Copy(part, file)
				if err != nil {
					return fmt.Errorf("failed to copy photo data: %w", err)
				}
				err = writer.Close()
				if err != nil {
					return fmt.Errorf("error closing writer: %w", err)
				}

				logoResp, err := client.UploadDirectoryLogoWithBodyWithResponse(cmd.Context(), resp.JSON200.ID, writer.FormDataContentType(), &requestBody)
				if err != nil {
					return err
				}

				if logoResp.StatusCode() != 204 {
					return fmt.Errorf("cannot save directory logo: %s", logoResp.Status())
				}
			}

			if jsonOutput {
				e := json.NewEncoder(cmd.OutOrStdout())
				e.SetIndent("", "\t")
				return e.Encode(resp.JSON200)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created a new directory: %s\n", resp.JSON200.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "You can run an agent for this directory with:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  export NAMETAG_AGENT_TOKEN=%q\n",
				lo.FromPtr(resp.JSON200.AgentToken))
			fmt.Fprintf(cmd.OutOrStdout(), "  nametag directory agent [provider]\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "See 'nametag directory agent --help' for more options.\n")
			return nil
		},
	}
	cmd.Flags().StringP("env", "e", "", "The environment to use for the directory")
	_ = cmd.MarkFlagRequired("env")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().StringP("logo", "l", "", "The logo for the directory")
	return cmd
}

func newDirAgentRegenerateTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate",
		Short: "Regenerate the directory agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			dirID, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}
			jsonOutput, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			resp, err := client.RegenerateDirectoryAgentTokenWithResponse(cmd.Context(), dirID)
			if err != nil {
				return err
			}
			if resp.StatusCode() != 200 {
				return fmt.Errorf("cannot regenerate directory agent token: %s", resp.Status())
			}

			if jsonOutput {
				e := json.NewEncoder(cmd.OutOrStdout())
				e.SetIndent("", "\t")
				return e.Encode(resp.JSON200)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "You can run an agent for this directory with:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  export NAMETAG_AGENT_TOKEN=%q\n", resp.JSON200.AgentToken)
			fmt.Fprintf(cmd.OutOrStdout(), "  nametag directory agent [provider]\n")
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "See 'nametag directory agent --help' for more options.\n")
			return nil
		},
	}
	cmd.Flags().StringP("dir", "d", "", "The ID of the directory you want to reconfigure")
	_ = cmd.MarkFlagRequired("dir")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	return cmd
}
