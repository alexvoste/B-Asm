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
	"sync/atomic"
	"testing"
)

func TestMPSCConcurrentDequeue(t *testing.T) {
	q := NewMPSC(256)
	const n = 1000
	for i := 0; i < n; i++ {
		if !q.Enqueue(i) {
			t.Fatal("enqueue failed")
		}
	}
	var got atomic.Int32
	var wg sync.WaitGroup
	workers := 8
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				v, ok := q.Dequeue()
				if !ok {
					break
				}
				got.Add(1)
				if v.(int) < 0 || v.(int) >= n {
					t.Errorf("unexpected value %v", v)
				}
			}
		}()
	}
	wg.Wait()
	if got.Load() != n {
		t.Fatalf("expected %d dequeued, got %d", n, got.Load())
	}
}

func TestMPSCDoesNotBlockWhenFull(t *testing.T) {
	q := NewMPSC(2)
	for i := 0; i < 1000; i++ {
		if !q.Enqueue(i) {
			t.Fatalf("enqueue %d failed", i)
		}
	}
	for i := 0; i < 1000; i++ {
		value, ok := q.Dequeue()
		if !ok || value.(int) != i {
			t.Fatalf("dequeue %d = (%v, %v)", i, value, ok)
		}
	}
}

func TestMPSCSteadyStateDoesNotAllocate(t *testing.T) {
	q := NewMPSC(16)
	allocs := testing.AllocsPerRun(1000, func() {
		if !q.Enqueue(1) {
			t.Fatal("enqueue failed")
		}
		if value, ok := q.Dequeue(); !ok || value.(int) != 1 {
			t.Fatalf("dequeue = (%v, %v)", value, ok)
		}
	})
	if allocs != 0 {
		t.Fatalf("steady-state allocations = %g, want 0", allocs)
	}
}

func BenchmarkMPSCSteadyState(b *testing.B) {
	q := NewMPSC(1024)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if !q.Enqueue(1) {
			b.Fatal("enqueue failed")
		}
		if _, ok := q.Dequeue(); !ok {
			b.Fatal("dequeue failed")
		}
	}
}
