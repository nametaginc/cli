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
	"io"
	"os"
)

// Writer returns an io.WriteCloser that writes to the given path,
// but only if the contents are different from the current contents.
func Writer(path string, opts ...writerOpt) io.WriteCloser {
	w := &writer{
		path: path,
		mode: 0644,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

type writerOpt func(w *writer)

// Mode sets the file mode of the file to something other than the default (0644)
func Mode(mode os.FileMode) writerOpt {
	return func(w *writer) {
		w.mode = mode
	}
}

type writer struct {
	path string
	mode os.FileMode
	buf  bytes.Buffer
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

func (w *writer) Close() error {
	current, _ := os.ReadFile(w.path)
	if bytes.Equal(current, w.buf.Bytes()) {
		st, err := os.Stat(w.path)
		if err != nil {
			return err
		}
		if st.Mode() != w.mode {
			if err := os.Chmod(w.path, w.mode); err != nil {
				return err
			}
		}
		return nil
	}
	return os.WriteFile(w.path, w.buf.Bytes(), w.mode)
}
