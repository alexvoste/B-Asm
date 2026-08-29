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

package lfqueue

import (
	"sync"
	"sync/atomic"
)

type node struct {
	val  any
	next atomic.Pointer[node]
}

var nodePool sync.Pool

type Queue struct {
	head atomic.Pointer[node]
	tail atomic.Pointer[node]
}

func init() {
	nodePool.New = func() any { return &node{} }
}

func New() *Queue {
	stub := nodePool.Get().(*node)
	stub.val = nil
	stub.next.Store(nil)
	q := &Queue{}
	q.head.Store(stub)
	q.tail.Store(stub)
	return q
}

func (q *Queue) Enqueue(v any) {
	n := nodePool.Get().(*node)
	n.val = v
	n.next.Store(nil)
	for {
		t := q.tail.Load()
		next := t.next.Load()
		if t == q.tail.Load() {
			if next == nil {
				if t.next.CompareAndSwap(nil, n) {
					q.tail.CompareAndSwap(t, n)
					return
				}
			} else {
				q.tail.CompareAndSwap(t, next)
			}
		}
	}
}

func (q *Queue) Dequeue() (any, bool) {
	for {
		h := q.head.Load()
		t := q.tail.Load()
		next := h.next.Load()
		if h == q.head.Load() {
			if h == t {
				if next == nil {
					return nil, false
				}
				q.tail.CompareAndSwap(t, next)
			} else {
				if q.head.CompareAndSwap(h, next) {
					v := next.val
					h.val = nil
					h.next.Store(nil)
					nodePool.Put(h)
					return v, true
				}
			}
		}
	}
}
