// Copyright 2023 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

//go:build !prod
// +build !prod

package thunks

import "time"

var defaultTime = time.Date(1981, 02, 03, 14, 15, 16, 17000, time.UTC)
var setTimeHooks []func(time.Time) // protected by timeNowMu

// SetTime sets the test clock to the specified time
func SetTime(t time.Time) {
	timeNowMu.Lock()
	defer timeNowMu.Unlock()
	timeNow = func() time.Time {
		return t
	}
	for _, hook := range setTimeHooks {
		hook(t)
	}
}

// AddSetTimeHook registers a hook function to be called each time the
// test clock is changed.
func AddSetTimeHook(hook func(time.Time)) {
	timeNowMu.Lock()
	defer timeNowMu.Unlock()
	setTimeHooks = append(setTimeHooks, hook)
}

// AdvanceNow moves the test clock forward by the specified amount
func AdvanceNow(d time.Duration) {
	SetTime(TimeNow().Add(d))
}

// ResetNow resets the test clock to the default time
func ResetNow() {
	SetTime(defaultTime)
}

func setUpTestForTime() {
	timeNowMu.Lock()
	defer timeNowMu.Unlock()
	setTimeHooks = nil
	timeNow = func() time.Time { return defaultTime }
}
