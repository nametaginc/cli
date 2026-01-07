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

	"github.com/nametaginc/cli/directory/dirldap"
	"github.com/nametaginc/cli/internal/config"
	"github.com/nametaginc/cli/internal/diragent"
)

func newDirAgentLDAPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ldap",
		Short: "Run the LDAP directory agent",
		Long: `Run the LDAP directory agent

The LDAP directory agent performs operations on behalf of Nametag such as listing accounts,
resetting passwords and unlocking accounts. Running a directory
agent allows you to shield your directory credentials from Nametag or customize the behavior
of already-supported directories.

All of the command line arguments can be configured in the nametag config file.
Command line arguments take precedence over config file values.

When invoked as a subcommand of 'nametag dir agent', the command runs as a worker, receiving
commands on stdin and sending responses to stdout. For example:

  nametag dir agent --agent-token <token> --command "nametag dir agent ldap"

For convenience, you can also invoke this command directly, which will cause it to perform
both the worker and the agent roles. For example, the following is equivalent to the above:

  nametag dir agent ldap --agent-token <token>

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

			cliConfig, err := config.ReadConfig(cmd)
			if err != nil {
				return err
			}

			ldapURL, err := cmd.Flags().GetString("ldap-url")
			if err != nil {
				return err
			}

			if ldapURL != "" {
				cliConfig.LDAPConfig.LDAPUrl = ldapURL
			}

			bindDn, err := cmd.Flags().GetString("bind-dn")
			if err != nil {
				return err
			}

			if bindDn != "" {
				cliConfig.LDAPConfig.BindDN = bindDn
			}

			bindPassword, err := cmd.Flags().GetString("bind-password")
			if err != nil {
				return err
			}

			if bindPassword != "" {
				cliConfig.LDAPConfig.BindPassword = bindPassword
			}

			baseDn, err := cmd.Flags().GetString("base-dn")
			if err != nil {
				return err
			}

			if baseDn != "" {
				cliConfig.LDAPConfig.BaseDN = baseDn
			}

			// If no pageSize is configured in config, we set a default
			if cliConfig.LDAPConfig.PageSize == 0 {
				cliConfig.LDAPConfig.PageSize = 250
			}

			provider := dirldap.Provider{
				Config: &cliConfig.LDAPConfig,
			}

			return diragent.RunWorker(cmd.Context(), &provider)
		},
	}
	cmd.Flags().String("agent-token", os.Getenv("NAMETAG_AGENT_TOKEN"), "Nametag directory agent authentication token ($NAMETAG_AGENT_TOKEN)")
	cmd.Flags().String("ldap-url", os.Getenv("LDAP_URL"), "ldap scheme URL")
	cmd.Flags().String("bind-dn", os.Getenv("BIND_DN"), "ldap bind DN")
	cmd.Flags().String("bind-password", os.Getenv("BIND_PASSWORD"), "ldap bind password")
	cmd.Flags().String("base-dn", os.Getenv("BASE_DN"), "ldap base DN")
	return cmd
}
