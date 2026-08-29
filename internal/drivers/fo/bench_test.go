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

package fo

import (
	"testing"
	"unsafe"
)

func BenchmarkFOSubmit(b *testing.B) {
	p := NewPool(4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := struct{}{}
		fn := func(arg unsafe.Pointer) error {
			_ = *(*struct{})(arg)
			return nil
		}
		t := Task{Fn: fn, Arg: unsafe.Pointer(&s)}
		for !p.Submit(t) {
		}
	}
	p.Stop()
}

func BenchmarkFOSubmitBatch(b *testing.B) {
	p := NewPool(4)
	b.ReportAllocs()
	batchSize := 64
	tasks := make([]Task, batchSize)
	for i := 0; i < batchSize; i++ {
		arg := new(struct{})
		fn := func(a unsafe.Pointer) error {
			_ = *(*struct{})(a)
			return nil
		}
		tasks[i] = Task{Fn: fn, Arg: unsafe.Pointer(arg)}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for !p.SubmitBatch(tasks) {
		}
	}
	p.Stop()
}
