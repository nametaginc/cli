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

	"github.com/nametaginc/cli/directory/dirbyid"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentBeyondIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "byid",
		Short: "Run the Beyond Identity directory agent",
		Long: `Run the Beyond Identity directory agent
The Beyond Identity directory agent performs operations on behalf of Nametag such as pulling in identities and groups from Beyond Identity.
Running a directory agent allows you to shield your directory credentials from Nametag or customize the behavior
of already-supported directories.

If you are using v0 API, you must specify a Beyond Identity URL, a client ID, and a client secret.
If you are using v1 API, you must specify a Beyond Identity URL, a client ID, a client secret,
a tenant ID, and a realm ID.

When invoked as a subcommand of 'nametag directory agent', the command runs as a worker, receiving
commands on stdin and sending responses to stdout.

For example:

v0 API:
    NAMETAG_AGENT_TOKEN="nametag-agent-token" nametag directory agent --command "NAMETAG_AGENT_TOKEN="nametag-agent-token" \
	BYID_URL="https://api.byndid.com/v2"\
	BYID_CLIENT_ID="client-id" \
	BYID_CLIENT_SECRET="client-secret" \
    nametag directory agent byid"

v1 API:
    NAMETAG_AGENT_TOKEN="nametag-agent-token" nametag directory agent --command "NAMETAG_AGENT_TOKEN="nametag-agent-token" \
	BYID_URL="https://api-us.beyondidentity.com/v1"\
	TENANT_ID="tenant-id" \
	REALM_ID="realm-id" \
	BYID_CLIENT_ID="client-id" \
	BYID_CLIENT_SECRET="client-secret" \
    nametag directory agent byid"

For convenience, you can also invoke this command directly, which will cause it to perform
both the worker and the agent roles. For example, the following is equivalent to the above:

v0 API:
	NAMETAG_AGENT_TOKEN="nametag-agent-token" \
	BYID_URL="https://api.byndid.com/v2" \
	BYID_CLIENT_ID="client-id" \
	BYID_CLIENT_SECRET="client-secret" \
    nametag directory agent byid"

v1 API:
	NAMETAG_AGENT_TOKEN="nametag-agent-token" \
	BYID_URL="https://api-us.beyondidentity.com/v1" \
	TENANT_ID="tenant-id" \
	REALM_ID="realm-id" \
	BYID_CLIENT_ID="client-id" \
	BYID_CLIENT_SECRET="client-secret" \
    nametag directory agent byid"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("byid-url")
			if err != nil {
				return err
			}
			if url == "" {
				return fmt.Errorf("flag url or environment variable $BYID_URL is required")
			}

			if url != "https://api-us.beyondidentity.com/v1" && url != "https://api.byndid.com/v2" {
				return fmt.Errorf("invalid url %s, must be https://api-us.beyondidentity.com/v1 or https://api.byndid.com/v2", url)
			}

			clientID, err := cmd.Flags().GetString("byid-client-id")
			if err != nil {
				return err
			}
			clientSecret, err := cmd.Flags().GetString("byid-client-secret")
			if err != nil {
				return err
			}
			if clientID == "" || clientSecret == "" {
				return fmt.Errorf("both byid-client-id and byid-client-secret are required")
			}

			// v1 API only
			var tenantID, realmID string
			if url == "https://api-us.beyondidentity.com/v1" {
				tenantID, err := cmd.Flags().GetString("tenant-id")
				if err != nil {
					return err
				}
				realmID, err := cmd.Flags().GetString("byid-realm-id")
				if err != nil {
					return err
				}
				if tenantID == "" || realmID == "" {
					return fmt.Errorf("both tenant-id and realm-id are required for v1 API")
				}
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

			provider := dirbyid.Provider{
				URL:          url,
				ClientID:     clientID,
				ClientSecret: clientSecret,
				TenantID:     &tenantID,
				RealmID:      &realmID,
			}
			return diragent.RunWorker(cmd.Context(), &provider)
		},
	}
	cmd.Flags().String("agent-token", os.Getenv("NAMETAG_AGENT_TOKEN"), "Nametag directory agent authentication token ($NAMETAG_AGENT_TOKEN)")
	cmd.Flags().String("byid-url", os.Getenv("BYID_URL"), "Your Beyond Identity URL ($BYID_URL)")
	cmd.Flags().String("byid-client-id", os.Getenv("BYID_CLIENT_ID"), "Your Beyond Identity Client ID ($BYID_CLIENT_ID)")
	cmd.Flags().String("byid-client-secret", os.Getenv("BYID_CLIENT_SECRET"), "Your Beyond Identity Client Secret ($BYID_CLIENT_SECRET)")
	cmd.Flags().String("tenant-id", os.Getenv("TENANT_ID"), "Your Beyond Identity Tenant ID ($TENANT_ID)")
	cmd.Flags().String("realm-id", os.Getenv("REALM_ID"), "Your Beyond Identity Realm ID ($REALM_ID)")
	return cmd
}
