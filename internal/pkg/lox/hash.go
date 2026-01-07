// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package lox

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"hash/crc32"
)

// SHA256 returns the SHA256 hash of buf as a byte slice
func SHA256(buf []byte) []byte {
	h := sha256.Sum256(buf)
	return h[:]
}

// SHA512 returns the SHA512 hash of buf as a byte slice
func SHA512(buf []byte) []byte {
	h := sha512.Sum512(buf)
	return h[:]
}

// CRC32 returns the CRC32 checksum of buf as an int
func CRC32(buf []byte) int {
	return int(crc32.ChecksumIEEE(buf))
}

// HMACSHA256 returns the HMAC-SHA256 hash of value using key
func HMACSHA256(key []byte, value []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(value)
	return h.Sum(nil)
}
