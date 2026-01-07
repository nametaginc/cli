// Copyright 2020 Nametag, Inc.
//
// All information contained herein is the property of Nametag, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package must

import "encoding/json"

// MarshalJSON returns the JSON encoding of v, or panics if v cannot be encoded.
func MarshalJSON(v interface{}) []byte {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return buf
}

// MarshalJSONPretty returns the JSON encoding of v, or panics if v cannot be encoded.
func MarshalJSONPretty(v interface{}) []byte {
	buf, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		panic(err)
	}
	return buf
}

// UnmarshalJSON parses JSON-encoded data, or panics if data cannot be decoded.
func UnmarshalJSON[T any](data []byte) T {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	return v
}
