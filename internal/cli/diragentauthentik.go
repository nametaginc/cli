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
	"fmt"
	"os"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/directory/dirauthentik"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentAuthentikCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authentik",
		Short: "Run the Authentik directory agent",
		Long: `Run the Authentik directory agent

The Authentik directory agent performs operations on behalf of Nametag such as listing accounts,
creating recovery links, and listing groups. Running a directory agent allows you to shield your
Authentik credentials from Nametag or customize the behavior of already-supported directories.

You must specify an Authentik URL and an Authentik API token.

When invoked as a subcommand of 'nametag directory agent', the command runs as a worker, receiving
commands on stdin and sending responses to stdout. For example:
  NAMETAG_AGENT_TOKEN="abcd" nametag directory agent --command "AUTHENTIK_TOKEN=... AUTHENTIK_URL=... nametag directory agent authentik"

For convenience, you can also invoke this command directly, which will cause it to perform both
the worker and the agent roles. For example, the following is equivalent to the above:
  NAMETAG_AGENT_TOKEN="abcd" \
  AUTHENTIK_TOKEN="..." \
  AUTHENTIK_URL="https://authentik.example.com" \
  nametag directory agent authentik
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("authentik-url")
			if err != nil {
				return err
			}
			if url == "" {
				return fmt.Errorf("flag authentik-url or environment variable $AUTHENTIK_URL is required")
			}

			token, err := cmd.Flags().GetString("authentik-token")
			if err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("flag authentik-token or environment variable $AUTHENTIK_TOKEN is required")
			}

			// we are not the worker, we are called as a top-level command, so run the agent,
			// passing the current command line as the command to run.
			if os.Getenv("NAMETAG_AGENT_WORKER") != "true" {
				agentToken, err := cmd.Flags().GetString("agent-token")
				if err != nil {
					return err
				}
				if agentToken == "" {
					agentToken = os.Getenv("NAMETAG_AGENT_TOKEN")
				}

				svc := diragent.Service{
					Server:    getServer(cmd),
					AuthToken: agentToken,
					Command:   shellquote.Join(os.Args...),
					Stderr:    cmd.ErrOrStderr(),
				}
				return svc.Run(cmd.Context())
			}

			provider := dirauthentik.Provider{
				URL:   url,
				Token: token,
			}
			return diragent.RunWorker(cmd.Context(), &provider)
		},
	}
	cmd.Flags().String("agent-token", os.Getenv("NAMETAG_AGENT_TOKEN"), "Nametag directory agent authentication token ($NAMETAG_AGENT_TOKEN)")
	cmd.Flags().String("authentik-url", os.Getenv("AUTHENTIK_URL"), "Your Authentik URL ($AUTHENTIK_URL)")
	cmd.Flags().String("authentik-token", os.Getenv("AUTHENTIK_TOKEN"), "Your Authentik API token ($AUTHENTIK_TOKEN)")
	return cmd
}
