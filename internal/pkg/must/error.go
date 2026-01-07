// Copyright 2021 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

// Package must contains utilities for testing
package must

import "io"

// NotFail panics if err is not nil
func NotFail(err error) {
	if err != nil {
		panic(err)
	}
}

// Close closes closer and panics if Close() returns a non-nil error
func Close(closer io.Closer) {
	NotFail(closer.Close())
}

// Return is useful for functions returning (value, error). It panics
// if err is not nil, otherwise it returns the value.
func Return[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
