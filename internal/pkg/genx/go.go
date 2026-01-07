// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package genx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Go returns a writer for formatted Go code.
func Go(path string) io.WriteCloser {
	w := &goWriter{path: path}
	fmt.Fprintf(w, `// Copyright 2024 Nametag Inc.
	//
	// All information contained herein is the property of Nametag Inc.. The
	// intellectual and technical concepts contained herein are proprietary, trade
	// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
	// and Foreign Patents, patents in process, and are protected by trade secret or
	// copyright law. Reproduction or distribution, in whole or in part, is
	// forbidden except by express written permission of Nametag, Inc.

	// Automatically generated. Do not edit.

`)
	return w
}

// goWriter is an io.Writer that emits Go code.
type goWriter struct {
	path   string
	buffer bytes.Buffer
}

func (w *goWriter) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}

func (w *goWriter) Close() error {
	{
		var formatted bytes.Buffer
		cmd := exec.Command("gofmt")
		cmd.Stdin = &w.buffer
		cmd.Stdout = &formatted
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			_ = os.MkdirAll(filepath.Dir(w.path), 0755)
			_ = os.WriteFile(w.path, w.buffer.Bytes(), 0644)
			return err
		}
		w.buffer = formatted
	}

	// run go imports too
	{
		var formatted bytes.Buffer
		cmd := exec.Command("go", "tool",
			"goimports",
			"-local", "github.com/nametaginc/nt")
		cmd.Stdin = &w.buffer
		cmd.Stdout = &formatted
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			_ = os.MkdirAll(filepath.Dir(w.path), 0755)
			_ = os.WriteFile(w.path, w.buffer.Bytes(), 0644)

			return err
		}
		w.buffer = formatted
	}

	// TODO(ross): run gogroup, but it doesn't support stdin

	existing, _ := os.ReadFile(w.path)
	if bytes.Equal(existing, w.buffer.Bytes()) {
		return nil
	}

	_ = os.MkdirAll(filepath.Dir(w.path), 0755)
	return os.WriteFile(w.path, w.buffer.Bytes(), 0644)
}
