// Copyright 2026 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package thunks

import (
	"time"

	"github.com/jpillora/backoff"
)

// DefaultBackoff returns a backoff with default max/min delays.
func DefaultBackoff() backoff.Backoff {
	return Backoff(0, 0)
}

// Backoff is a thunk for creating backoff configurations.
var Backoff = normalBackoff

func normalBackoff(minDuration time.Duration, maxDuration time.Duration) backoff.Backoff {
	return backoff.Backoff{
		Min: minDuration,
		Max: maxDuration,
	}
}
