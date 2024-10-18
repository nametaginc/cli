// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package main

import (
	"fmt"
	"os"

	"github.com/nametaginc/cli/internal/cli"
)

func main() {
	if err := cli.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
