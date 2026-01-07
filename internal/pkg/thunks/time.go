// Copyright 2020 Nametag, Inc.
//
// All information contained herein is the property of Nametag, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

// Package thunks contains various thunk utilities
package thunks

import (
	"sync"
	"time"
)

// TimeNow returns the current time, but may be replaced in tests
func TimeNow() time.Time {
	timeNowMu.RLock()
	defer timeNowMu.RUnlock()
	return timeNow()
}

var (
	timeNowMu sync.RWMutex // protects timeNow (and setTimeHooks in test code)
	timeNow   = func() time.Time { return time.Now().UTC() }
)
