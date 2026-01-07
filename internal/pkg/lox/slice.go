// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package lox

import "github.com/samber/lo"

// Filter iterates over elements of collection, returning an array of all elements predicate returns truthy for.
// Play: https://go.dev/play/p/Apjg3WeSi7K
func Filter[T any, Slice ~[]T](collection Slice, predicate func(item T) bool) Slice {
	return lo.Filter(collection, func(item T, index int) bool {
		return predicate(item)
	})
}

// Map manipulates a slice and transforms it to a slice of another type.
// Play: https://go.dev/play/p/OkPcYAhBo0D
func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	return lo.Map(collection, func(item T, index int) R {
		return iteratee(item)
	})
}

// Last returns the last element of a collection or zero value if empty.
func Last[T any](collection []T) T {
	return lo.LastOrEmpty(collection)
}
