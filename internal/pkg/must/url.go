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
	"net/url"
)

// ParseURL parses the specified URL, panicing if it is not valid
func ParseURL(s string) *url.URL {
	rv, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return rv
}
