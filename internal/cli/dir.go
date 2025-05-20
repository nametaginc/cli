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

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/api"
)

func newDirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dir",
		Aliases: []string{"directory", "directories", "dirs"},
		Short:   "Directory tools",
	}
	cmd.AddCommand(newDirListCmd())
	cmd.AddCommand(newDirGetCmd())
	cmd.AddCommand(newDirAgentCmd())

	return cmd
}
func newDirListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := client.ListDirectoriesWithResponse(cmd.Context())
			if err != nil {
				return err
			}
			if resp.StatusCode() != 200 {
				return fmt.Errorf("%s", resp.Status())
			}

			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()

			tbl := table.New("ID", "Env", "Kind", "Name")
			tbl.
				WithWriter(cmd.OutOrStdout()).
				WithHeaderFormatter(headerFmt).
				WithFirstColumnFormatter(columnFmt)
			for _, dir := range lo.FromPtr(resp.JSON200).Directories {
				tbl.AddRow(dir.ID, dir.Env, dir.Kind, dir.Name)
			}
			tbl.Print()
			return nil
		},
	}
}

func newDirGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [dir]",
		Short: "show an directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := client.GetDirectoryWithResponse(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if resp.StatusCode() >= 400 {
				return fmt.Errorf("%s", resp.Status())
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", resp.JSON200.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Env: %s\n", resp.JSON200.Env)
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", resp.JSON200.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Kind: %s\n", resp.JSON200.Kind)

			printPolicy := func(policy api.RecoveryPolicyRules) {
				for _, groupPolicy := range policy.Groups {
					fmt.Fprintf(cmd.OutOrStdout(), "  - Group: %s (%s)\n", groupPolicy.Group.Name, groupPolicy.Group.DirectoryImmutableIdentifier)
					fmt.Fprintf(cmd.OutOrStdout(), "    Policy: %s\n", groupPolicy.Policy)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  - Default: %s\n", policy.Default)
			}

			if t := resp.JSON200.LastSyncStartedAt; t != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "LastSyncStartedAt: %s\n", *t)
			}
			if t := resp.JSON200.LastSyncCompletedAt; t != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "LastSyncCompletedAt: %s\n", *t)
			}
			if c := resp.JSON200.Count; c != nil && *c > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Count: %d\n", *c)
			}
			if v := resp.JSON200.NeedsReconnect; v != nil && *v {
				fmt.Fprintf(cmd.OutOrStdout(), "NeedsReconnect: true\n")
			}
			if v := resp.JSON200.SyncRunning; v {
				fmt.Fprintf(cmd.OutOrStdout(), "SyncRunning: true\n")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "AuthenticatePolicy:\n")
			printPolicy(resp.JSON200.AuthenticatePolicy)
			fmt.Fprintf(cmd.OutOrStdout(), "MFAPolicy:\n")
			printPolicy(resp.JSON200.MfaPolicy)
			fmt.Fprintf(cmd.OutOrStdout(), "PasswordPolicy:\n")
			printPolicy(resp.JSON200.PasswordPolicy)
			fmt.Fprintf(cmd.OutOrStdout(), "UnlockPolicy:\n")
			printPolicy(resp.JSON200.UnlockPolicy)
			fmt.Fprintf(cmd.OutOrStdout(), "TemporaryAccessPassPolicy:\n")
			printPolicy(resp.JSON200.TemporaryAccessPassPolicy)

			return nil
		},
	}
}
