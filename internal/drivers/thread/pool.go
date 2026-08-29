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

package thread

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	ch "github.com/forgezero-cli/ForgeZero/internal/drivers/chan"
)

type Task func()

type worker struct {
	q *ch.MPSC
}

type Pool struct {
	workers []*worker
	next    atomic.Uint64
	stop    atomic.Uint32
	wg      sync.WaitGroup
}

func NewPool(size int) *Pool {
	if size <= 0 {
		size = 1
	}
	p := &Pool{workers: make([]*worker, size)}
	for i := 0; i < size; i++ {
		q := ch.NewMPSC(1024)
		w := &worker{q: q}
		p.workers[i] = w
		p.wg.Add(1)
		go func(w *worker) {
			defer p.wg.Done()
			idle := uint32(0)
			for p.stop.Load() == 0 {
				if v, ok := w.q.Dequeue(); ok {
					idle = 0
					if t, ok2 := v.(Task); ok2 {
						t()
					}
					continue
				}
				idle++
				if idle < 64 {
					runtime.Gosched()
					continue
				}
				time.Sleep(time.Microsecond << (idle & 7))
			}
		}(w)
	}
	return p
}

func (p *Pool) Submit(t Task) bool {
	if p.stop.Load() == 1 {
		return false
	}
	idx := int(p.next.Add(1)-1) % len(p.workers)
	for !p.workers[idx].q.Enqueue(t) {
		runtime.Gosched()
	}
	return true
}

func (p *Pool) Stop() {
	p.stop.Store(1)
	p.wg.Wait()
}
