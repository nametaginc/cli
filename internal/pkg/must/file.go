// Copyright 2020 Nametag, Inc.
//
// All information contained herein is the property of Nametag, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package must

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ReadFile returns the contents of the specified file.
func ReadFile(filename string) []byte {
	buf, err := os.ReadFile(filename) //#nosec G304 //we control the input here
	if err != nil {
		panic(err)
	}
	return buf
}

// ReadFileFS returns the contents of the specified file.
func ReadFileFS(fs fs.FS, filename string) []byte {
	f := Return(fs.Open(filename))
	defer Close(f)
	return Return(io.ReadAll(f))
}

// WriteFile writes the specified file
func WriteFile(filename string, content []byte) {
	_ = os.MkdirAll(filepath.Dir(filename), 0700)
	err := os.WriteFile(filename, content, 0644) //#nosec G306 // permissions are reasonable for this use case
	if err != nil {
		panic(err)
	}
}
