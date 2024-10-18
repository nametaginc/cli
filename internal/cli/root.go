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
	_ "embed"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed "VERSION"
var versionBuf []byte

var Version = strings.TrimSpace(string(versionBuf))

var Root = &cobra.Command{
	Use:           "nametag",
	Short:         "Nametag command line interface",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func subcmd(parent *cobra.Command, child *cobra.Command) *cobra.Command {
	parent.AddCommand(child)
	return child
}
