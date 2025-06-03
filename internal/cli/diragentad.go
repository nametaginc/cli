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
	"os"

	"github.com/kballard/go-shellquote"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/directory/dirad"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentADCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ad",
		Short: "Run the AD directory agent",
		Long: `Run the AD directory agent

The AD directory agent performs operations on behalf of Nametag such as listing accounts,
resetting passwords and unlocking accounts. Running a directory
agent allows you to shield your directory credentials from Nametag or customize the behavior
of already-supported directories.

Currently only supports running on the Windows platform.

It uses powershell to communicate with the AD server.
The current user must have permission to run powershell commands with sufficient privileges.
The ActiveDirectory module must be installed.

When invoked as a subcommand of 'nametag dir agent', the command runs as a worker, receiving
commands on stdin and sending responses to stdout. For example:

  nametag dir agent --agent-token <token> --command "nametag dir agent ad"

For convenience, you can also invoke this command directly, which will cause it to perform
both the worker and the agent roles. For example, the following is equivalent to the above:

  nametag dir agent ad --agent-token <token>

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// we are not the worker, we are called as a top-level command, so run the agent,
			// passing the current command line as the command to run.
			if os.Getenv("NAMETAG_AGENT_WORKER") != "true" {
				agentToken, err := cmd.Flags().GetString("agent-token")
				if err != nil {
					return err
				}

				svc := diragent.Service{
					Server:    getServer(cmd),
					AuthToken: agentToken,
					Command:   shellquote.Join(os.Args...),
					Stderr:    cmd.ErrOrStderr(),
				}
				return svc.Run(cmd.Context())
			}

			provider := dirad.Provider{}

			return diragent.RunWorker(cmd.Context(), &provider)
		},
	}
	cmd.Flags().String("agent-token", os.Getenv("NAMETAG_AGENT_TOKEN"), "Nametag directory agent authentication token ($NAMETAG_AGENT_TOKEN)")
	return cmd
}
