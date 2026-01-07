// Copyright 2025 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package lox

import "iter"

// Index returns an iterator over the elements of iter, yielding the index and the element.
func Index[T any](iter iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for item := range iter {
			if !yield(i, item) {
				return
			}
			i++
		}
	}
}

// Chunk breaks the elements of iter into chunks of size,
// returning an iterator over the chunks.
func Chunk[T any](size int, iter iter.Seq[T]) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		batch := make([]T, 0, size)
		for item := range iter {
			batch = append(batch, item)
			if len(batch) == size {
				if !yield(batch) {
					return
				}
				batch = make([]T, 0, size)
			}
		}
		if len(batch) > 0 {
			yield(batch)
		}
	}
}
