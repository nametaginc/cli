// Copyright 2025 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package genx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nametaginc/cli/internal/pkg/must"
	"github.com/nametaginc/cli/internal/pkg/thunks"
)

// Log emits a log message that memorializes the start of the code generation process,
// returning a function that, when called, will log the completion time of the generation.
//
// Use it like:
//
//	func main() {
//		defer genx.Log()()
//		// ...
//	}
func Log() func() {
	pwd := must.Return(os.Getwd())
	pwd = must.Return(filepath.Rel(must.Return(SourceRoot()), pwd))
	startTime := thunks.TimeNow()

	fmt.Printf("generate: %s\n", pwd)
	return func() {
		fmt.Printf("generate: %s %s\n", pwd, thunks.TimeNow().Sub(startTime).Round(time.Millisecond).String())
	}
}
