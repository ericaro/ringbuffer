// Copyright 2014 @ericaro. All rights reserved.
// Use of this source code is governed by a Apache License, Version 2.0.

// Package ringbuffer provides "Ring" struct: a ring buffer.
//
//
// A ring buffer is a data structure that uses a single, fixed-size buffer as if it were connected end-to-end.
// in http://en.wikipedia.org/wiki/Circular_buffer .
//
//
// Basic operations on a Ring are:
//   Add: add a value to the head.
//   Remove: remove value from the tail.
//   Get : read values from the Ring
//
// More advanced operations are:
//   Push: Add and Remove at once. It does not consume any extra memory
//   AddAll: Add several values in Bulk.
// 		SetCapacity: increase this buffer capacity (preserving its size)
//
//
package ringbuffer

import (
	"errors"
	"sync"
)

var (
	EmptyError = errors.New("empty ring buffer")
	FullError  = errors.New("full ring buffer")
)

//Ring is a basic implementation of a circular buffer http://en.wikipedia.org/wiki/Circular_buffer
// or Ring Buffer
type Ring struct {
	lock       sync.RWMutex
	head, size int
	buf        []interface{}
}

//New creates a new, empty ring buffer.
func New(capacity int) (b *Ring) {
	return &Ring{
		buf:  make([]interface{}, capacity),
		head: -1,
	}
}

//Add 'val' at the Ring's head, it also increases its size.
//If the capacity is exhausted (size == capacity) an error is returned.
func (b *Ring) Add(val interface{}) error {
	if b.size >= len(b.buf) {
		return FullError
	}
	b.lock.Lock()
	defer b.lock.Unlock()

	next := Next(1, b.head, len(b.buf))
	b.buf[next] = val
	b.head = next
	b.size++ // increase the inner size
	return nil
}

// AddAll add all values to the Ring's head.
// Behave like looping over Add() method, except that
// it uses bulk operations.
// If you try to add too much values, an error is returned and no value is actually added.
func (b *Ring) AddAll(values ...interface{}) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.size+len(values) > len(b.buf) {
		return FullError
	}

	//alg: add as much as possible in a single copy, and repeat until exhaustion

	for len(values) > 0 {
		// is all about slicing right
		//
		// the extra space to add value too can be either
		// from next position, to the end of the buffer
		// or from the beginning to the tail of the ring

		next := Next(1, b.head, len(b.buf))

		//tail := Index(-1, b.head, b.size, len(b.buf)

		// tail is the absolute index of the buffer tail
		// next is the absolute index of the buffer head+1
		//I want to write as much as possible after next.
		// so it's either till the end, or till the tail
		// if the ring's tail is behind we can use the slice from next to the end
		// if the ring's tail is ahead we can't use the whole slice.
		// BUT, we know that there is enough room left, so we don't care if the slice is "too" big
		// Therefore, instead of dealing with all the cases
		// we always use:
		tgt := b.buf[next:]
		// a slice of the buffer, that we'll used to write into

		//we copy as much as possible.
		n := copy(tgt, values) //n is the number of copied values

		if n == 0 { // could not write ! the buf is exhausted
			panic(FullError) // because we have tested this case before,I'd rather  panic than infinite loop
		}

		// we adjust local variables (latest has moved, and so has size)
		b.head = Next(n, b.head, len(b.buf))
		b.size += n // increase the inner size

		// we remove from the source, the value copied.
		values = values[n:]
	}
	return nil

}

//Remove 'count' items at the Ring's tail.
// If count is greater than the Ring's size, the Ring is set to empty.
func (b *Ring) Remove(count int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if count <= 0 {
		return
	}

	b.size -= count
	if b.size <= 0 {
		b.size = 0
		b.head = -1 //small trick to mark as empty
	}
	return
}

//SetCapacity tries to set the ring's capacity.
// The Ring content is not altered as a consequence of this operation,
// therefore the final capacity is at least equal to the Ring's size.
func (b *Ring) SetCapacity(capacity int) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if capacity < b.size {
		capacity = b.size
	}
	if capacity == len(b.buf) { //nothing to be done
		return
	}

	nbuf := make([]interface{}, capacity)

	// now that the new capacity is enough we just copy down the buffer

	//there are only two cases:
	// either the values are contiguous, then they goes from
	// tail to head
	// or there are splitted in two:
	// tail to buffer's end
	// 0 to head.

	head := b.head
	tail := Index(-1, head, b.size, len(b.buf))

	// we are not going to copy the buffer in the same state (absolute position of head and tail)
	// instead, we are going to select the simplest solution.
	if tail < head { //data is in one piece
		copy(nbuf, b.buf[tail:head+1])
	} else { //two pieces
		//copy as much as possible to the end of the buf
		n := copy(nbuf, b.buf[tail:])
		//and then from the beginning
		copy(nbuf[n:], b.buf[:head+1])
	}
	b.buf = nbuf
	b.head = b.size - 1
	return
}

//Capacity is the max size permitted
func (b *Ring) Capacity() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.buf)
}

//Size returns the Ring's size.
func (b *Ring) Size() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.size
}

//Push 'value' into the ring and discard the oldest one.
func (b *Ring) Push(value interface{}) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if len(b.buf) == 0 || b.size == 0 { // nothing to do
		return
	}
	next := Next(1, b.head, len(b.buf))
	b.buf[next] = value
	b.head = next
	// note that the oldest is auto pruned, when size== capacity, but with the size attribute we know it has been discarded
}

//Get returns the value in the ring.
//
// 'Get(0)' retreive the head
// 'Get(size-1)' is the oldest
// 'Get(-1)' is the oldest too.
func (b *Ring) Get(i int) (interface{}, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if b.size == 0 {
		return 0, EmptyError
	}
	position := Index(i, b.head, b.size, len(b.buf))
	return b.buf[position], nil

	// this pos might be negative too, so I need to make it positive
	// note that the oldest is auto pruned, when size== capacity, but with the size attribute we know it has been discarded
}

// Next computes the next index for a ring buffer
func Next(i, latest, capacity int) int {
	n := (latest + i) % capacity
	if n < 0 {
		n += capacity
	}
	return n
}

//Index computes absolute position of a ring buffer index.
//
// i, is the ring's index.
//
// head, is the absolute index of the ring's head
//
// size, is the ring' size
//
// capacity is the buffer's capacity.
//
func Index(i, head, size, capacity int) int {
	// size=0 is a failure.
	if size == 0 {
		return -1
	}
	// first fold i values into  ]-size , size[
	i = i % size
	// then translate negative parts
	if i < 0 {
		i += size
	}

	// this way -1 is interpreted as size-1 etc.

	// now I've got the real i>=0
	// actual theoretical index is simply
	// last write minus the required offset.
	// last write is lastest
	// offset is i, because i==0 means exactly the last written.
	//
	pos := head - i

	//pos might be negative. this is the actual index in the ring buffer.
	// if head = 0, previous read is at len(buf)-1
	// if head == 0 (and i was zero), pos=-1 (as the above calculation)
	//so this is the same as before, negative indexes are added the actual size
	for pos < 0 {
		pos += capacity
	}

	// yehaa, pos is the head position.
	return pos
}
