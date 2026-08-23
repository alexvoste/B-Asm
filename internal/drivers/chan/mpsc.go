/*
 *   Copyright (c) 2026 qecko-labs
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
)

type MPSC struct {
	mu    sync.Mutex
	head  uint64
	tail  uint64
	slots []any
}

func NewMPSC(capPow2 int) *MPSC {
	cap := uint64(1)
	for cap < uint64(capPow2) {
		cap <<= 1
	}
	return &MPSC{slots: make([]any, cap)}
}

func (q *MPSC) Enqueue(v any) bool {
	if q == nil {
		return false
	}
	q.mu.Lock()
	if q.tail-q.head == uint64(len(q.slots)) {
		q.grow()
	}
	q.slots[q.tail%uint64(len(q.slots))] = v
	q.tail++
	q.mu.Unlock()
	return true
}

func (q *MPSC) Dequeue() (any, bool) {
	if q == nil {
		return nil, false
	}
	q.mu.Lock()
	if q.head == q.tail {
		q.mu.Unlock()
		return nil, false
	}
	index := q.head % uint64(len(q.slots))
	v := q.slots[index]
	q.slots[index] = nil
	q.head++
	q.mu.Unlock()
	return v, true
}

func (q *MPSC) grow() {
	oldSize := uint64(len(q.slots))
	newSlots := make([]any, oldSize*2)
	for index := q.head; index < q.tail; index++ {
		newSlots[index-q.head] = q.slots[index%oldSize]
	}
	q.slots = newSlots
	q.tail -= q.head
	q.head = 0
}
