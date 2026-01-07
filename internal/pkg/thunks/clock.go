// Copyright 2023 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package thunks

import "time"

// Clock implements the clockwork.Clock interface
type Clock struct {
}

// After is time.After
func (fc Clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Sleep is time.Sleep
func (fc Clock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// Now returns TimeNow(), the fake/internal time
func (fc Clock) Now() time.Time {
	return TimeNow()
}

// Since returns the time between t and the fake/internal time returned by TimeNow()
func (fc Clock) Since(t time.Time) time.Duration {
	return TimeNow().Sub(t)
}
