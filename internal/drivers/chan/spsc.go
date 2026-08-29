/*
 *   Copyright (c) 2026 forgezero-cli
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package ch

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type SPSC struct {
	cap   uint64
	mask  uint64
	head  uint64
	tail  uint64
	slots []unsafe.Pointer
	pool  sync.Pool
}

func NewSPSC(capPow2 int) *SPSC {
	cap := uint64(1)
	for cap < uint64(capPow2) {
		cap <<= 1
	}
	s := &SPSC{cap: cap, mask: cap - 1, slots: make([]unsafe.Pointer, cap)}
	s.pool.New = func() any { return new(interface{}) }
	return s
}

func (s *SPSC) Enqueue(v any) bool {
	p := s.pool.Get().(*interface{})
	*p = v
	tail := atomic.LoadUint64(&s.tail)
	head := atomic.LoadUint64(&s.head)
	if tail-head >= s.cap {
		s.pool.Put(p)
		return false
	}
	idx := tail & s.mask
	atomic.StorePointer(&s.slots[idx], unsafe.Pointer(p))
	atomic.StoreUint64(&s.tail, tail+1)
	return true
}

func (s *SPSC) Dequeue() (any, bool) {
	head := atomic.LoadUint64(&s.head)
	tail := atomic.LoadUint64(&s.tail)
	if head >= tail {
		return nil, false
	}
	idx := head & s.mask
	p := atomic.LoadPointer(&s.slots[idx])
	if p == nil {
		return nil, false
	}
	atomic.StorePointer(&s.slots[idx], nil)
	atomic.StoreUint64(&s.head, head+1)
	v := *(*interface{})(p)
	s.pool.Put((*interface{})(p))
	return v, true
}
