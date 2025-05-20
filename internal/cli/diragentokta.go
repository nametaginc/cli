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

	"github.com/nametaginc/cli/directory/dirokta"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentOktaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "okta",
		Short: "Run the Okta directory agent",
		Long: `Run the Okta directory agent
The Okta directory agent performs operations on behalf of Nametag such as listing accounts, 
resetting passwords, resetting MFA, and unlocking accounts. Running a directory 
agent allows you to shield your directory credentials from Nametag or customize the behavior 
of already-supported directories.
You must specify an Okta URL and either (1) an Okta API token or (2) an Okta client ID and secret.
When invoked as a subcommand of 'nametag directory agent', the command runs as a worker, receiving
commands on stdin and sending responses to stdout. For example:
    NAMETAG_AGENT_TOKEN="abcd" nametag directory agent --command "NAMETAG_AGENT_TOKEN="abcd" \
	OKTA_TOKEN="1234567890" \
	OKTA_URL="https://example.okta.com" \
    nametag directory agnet okta"
For convenience, you can also invoke this command directly, which will cause it to perform
both the worker and the agent roles. For example, the following is equivalent to the above:
	NAMETAG_AGENT_TOKEN="abcd" \
	OKTA_TOKEN="1234567890" \
	OKTA_URL="https://example.okta.com" \
    nametag directory agnet okta
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("okta-url")
			if err != nil {
				return err
			}
			if url == "" {
				return fmt.Errorf("flag url or environment variable $OKTA_URL is required")
			}

			token, err := cmd.Flags().GetString("okta-token")
			if err != nil {
				return err
			}
			clientID, err := cmd.Flags().GetString("okta-client-id")
			if err != nil {
				return err
			}
			clientSecret, err := cmd.Flags().GetString("okta-client-secret")
			if err != nil {
				return err
			}

			if token != "" {
				// ok
			} else if clientID != "" && clientSecret != "" {
				// ok
			} else {
				return fmt.Errorf("at least one of okta-token or both okta-client-id and okta-client-secret are required")
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

			provider := dirokta.Provider{
				URL:          url,
				Token:        token,
				ClientID:     clientID,
				ClientSecret: clientSecret,
			}
			return diragent.RunWorker(cmd.Context(), &provider)
		},
	}
	cmd.Flags().String("agent-token", os.Getenv("NAMETAG_AGENT_TOKEN"), "Nametag directory agent authentication token ($NAMETAG_AGENT_TOKEN)")
	cmd.Flags().String("okta-url", os.Getenv("OKTA_URL"), "Your Okta URL ($OKTA_URL)")
	cmd.Flags().String("okta-token", os.Getenv("OKTA_TOKEN"), "Your Okta API key ($OKTA_TOKEN)")
	cmd.Flags().String("okta-client-id", os.Getenv("OKTA_CLIENT_ID"), "Your Okta Client ID ($OKTA_CLIENT_ID)")
	cmd.Flags().String("okta-client-secret", os.Getenv("OKTA_CLIENT_SECRET"), "Your Okta Client Secret ($OKTA_CLIENT_SECRET)")
	return cmd
}
