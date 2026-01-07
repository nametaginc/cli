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
	"crypto/rand"
	"io"
	"math/big"

	"github.com/nametaginc/cli/internal/pkg/thunks"
)

// ReadRandomBytes returns a random buffer of length n, or panics
// if bytes cannot be obtained.
//
// Uses thunks.RandReader as the source of randomness.
func ReadRandomBytes(n int) []byte {
	rv := make([]byte, n)
	_, err := io.ReadFull(thunks.RandReader, rv)
	if err != nil {
		panic(err)
	}
	return rv
}

// ReadRandomInteger returns an integer between [0, maxValue). It panics if max = 0
// or if random bytes cannot be read.
func ReadRandomInteger(maxValue uint64) uint64 {
	big.NewInt(0).SetUint64(maxValue)
	n, err := rand.Int(thunks.RandReader, big.NewInt(0).SetUint64(maxValue))
	if err != nil {
		panic(err)
	}
	return n.Uint64()
}
