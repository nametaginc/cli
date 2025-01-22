// Copyright 2025 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package jsonx

import (
	"encoding/json"
	"time"
)

// Duration is a time.Duration the serializes as a string in JSON
// using the go time.Duration string format.
type Duration time.Duration

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Duration(d).String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	// Note: Prior to 2025-01-17, we used time.Duration as the time for
	// public JSON fields. time.Duration serializes as nanoseconds. So we
	// don't break things, we accept the old format.
	var nanoseconds int64
	if json.Unmarshal(data, &nanoseconds) == nil {
		*d = Duration(nanoseconds)
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	v, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}
