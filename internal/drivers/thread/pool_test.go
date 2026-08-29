/*
 *   Copyright (c) 2026 forgezero-cli
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package thread

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolSubmit(t *testing.T) {
	p := NewPool(2)
	var c int32
	for i := 0; i < 100; i++ {
		p.Submit(func() { atomic.AddInt32(&c, 1) })
	}
	time.Sleep(50 * time.Millisecond)
	p.Stop()
	if atomic.LoadInt32(&c) != 100 {
		t.Fatalf("expected 100 got %d", c)
	}
}
