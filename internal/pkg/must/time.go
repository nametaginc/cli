// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package must

import "time"

// ParseTime returns the time.Time value represented by the string s
// in RFC3339 format.
func ParseTime(s string) time.Time {
	return Return(time.Parse(time.RFC3339, s))
}
