// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var Env = subcmd(Root, &cobra.Command{
	Use:     "env",
	Aliases: []string{"environment", "environments", "envs"},
	Short:   "Commands for working with environments",
})

var _ = subcmd(Env, &cobra.Command{
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
			WithHeaderFormatter(headerFmt).
			WithFirstColumnFormatter(columnFmt)
		for _, env := range lo.FromPtr(resp.JSON200).Envs {
			tbl.AddRow(env.ID, env.Name, env.PublicName)
		}
		tbl.Print()
		return nil
	},
})

var _ = subcmd(Env, &cobra.Command{
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

		fmt.Printf("ID: %s\n", resp.JSON200.ID)
		fmt.Printf("Name: %s\n", resp.JSON200.Name)
		fmt.Printf("PublicName: %s\n", resp.JSON200.PublicName)
		fmt.Printf("LogoURL: %s\n", resp.JSON200.LogoURL)
		return nil
	},
})
