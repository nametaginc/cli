// Copyright 2020 Nametag, Inc.
//
// All information contained herein is the property of Nametag, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package must

import (
	"encoding/hex"
)

// DecodeHex decodes the hex-encoded string in encoded, or panics if
// it is not a valid string.
func DecodeHex(encoded string) []byte {
	rv, err := hex.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	return rv
}
