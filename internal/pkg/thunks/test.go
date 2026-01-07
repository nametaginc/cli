// Copyright 2020 Nametag, Inc.
//
// All information contained herein is the property of Nametag, Inc. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

//go:build !prod
// +build !prod

package thunks

import (
	"fmt"
	"io"
	mathrand "math/rand"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// TestReader returns bytes from Next whenever asked, unless the
// number of bytes requested is 1, in which case it always returns 0x00.
//
// This is because in go >= 1.11 the crypto libraries non-deterministically read
// a single byte from the PRNG in order to prevent people doing what we do and
// relying on the internal details of cryptographic operations.
//
// For our purposes, having deterministic crypto in tests is super handy, and
// we have mechanisms for updating our expectations. So, while it might be better
// if we didn't rely on the crypto/* internals, on balance, the convenience of
// fixed test data outweighs the uglyness of this hack.
//
// You might ask why we don't use a zero-reader as suggested by the commit
// message. The reason is that rsa key generation seems to spin forever when given a
// random source that always returns 0.
//
// [1] https://github.com/golang/go/commit/6269dcdc24d74379d8a609ce886149811020b2cc
type TestReader struct {
	Next io.Reader
	mu   sync.Mutex
}

func (z *TestReader) Read(dst []byte) (n int, err error) {
	if len(dst) == 1 {
		dst[0] = 0
		return 1, nil
	}

	z.mu.Lock()
	defer z.mu.Unlock()
	return z.Next.Read(dst)
}

// SetUpTest sets RandReader to be a deterministic PRNG, and
// TimeNow to return a fixed date
func SetUpTest() {
	// We need to be in UTC. This incantation should silence the race detector
	if time.Local != time.UTC {
		time.Local = time.UTC
	}
	RandReader = &TestReader{Next: mathrand.New(mathrand.NewSource(0))} //#nosec G404  // used for testing, not intended to be secure

	setUpTestForTime()

	GenerateFromPassword = func(password []byte, cost int) ([]byte, error) {
		return []byte(fmt.Sprintf("bcrypt(%s)", password)), nil
	}
	CompareHashAndPassword = func(hashedPassword, password []byte) error {
		pwString := fmt.Sprintf("bcrypt(%s)", string(password))
		hashString := string(hashedPassword)
		if hashString != pwString {
			return bcrypt.ErrMismatchedHashAndPassword
		}
		return nil
	}
}
