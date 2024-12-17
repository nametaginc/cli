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
)

func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "env",
		Aliases: []string{"environment", "environments", "envs"},
		Short:   "Commands for working with environments",
	}
	cmd.AddCommand(newEnvListCmd())
	cmd.AddCommand(newEnvGetCmd())
	return cmd
}

func newEnvListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := client.ListEnvsWithResponse(cmd.Context())
			if err != nil {
				return err
			}
			if resp.StatusCode() != 200 {
				return fmt.Errorf("%s", resp.Status())
			}

			headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
			columnFmt := color.New(color.FgYellow).SprintfFunc()

			tbl := table.New("ID", "Name", "Public Name")
			tbl.
				WithWriter(cmd.OutOrStdout()).
				WithHeaderFormatter(headerFmt).
				WithFirstColumnFormatter(columnFmt)
			for _, env := range lo.FromPtr(resp.JSON200).Envs {
				tbl.AddRow(env.ID, env.Name, env.PublicName)
			}
			tbl.Print()
			return nil
		},
	}
	return cmd
}

func newEnvGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [env]",
		Short: "show an environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := NewAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := client.GetEnvWithResponse(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if resp.StatusCode() >= 400 {
				return fmt.Errorf("%s", resp.Status())
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\n", resp.JSON200.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", resp.JSON200.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "PublicName: %s\n", resp.JSON200.PublicName)
			fmt.Fprintf(cmd.OutOrStdout(), "LogoURL: %s\n", resp.JSON200.LogoURL)
			return nil
		},
	}
	return cmd
}
