// Copyright 2025 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package lox

// Elide returns a string that is at most maxLen characters long.
// If the string is longer than maxLen, it will be truncated and "..." will be
// appended to the end. If maxLen is less than 3, it will be set to 3.
// If the string is shorter than or equal to maxLen, it will be returned
func Elide(s string, maxLen int) string {
	elipsis := []rune("…")
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < len(elipsis) {
		maxLen = len(elipsis)
	}
	return string(runes[0:maxLen-len(elipsis)]) + "…"
}
