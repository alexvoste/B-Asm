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

package syncx

import (
	"runtime"
	"sync/atomic"
	"time"
)

type SpinLock struct {
	flag uint32
}

func NewSpinLock() *SpinLock { return &SpinLock{} }

func (s *SpinLock) Lock() {
	for spin := 0; !atomic.CompareAndSwapUint32(&s.flag, 0, 1); spin++ {
		if spin < 8 {
			runtime.Gosched()
			continue
		}
		if spin < 16 {
			time.Sleep(time.Microsecond)
			continue
		}
		time.Sleep(10 * time.Microsecond)
	}
}

func (s *SpinLock) Unlock() {
	atomic.StoreUint32(&s.flag, 0)
}

func (s *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32(&s.flag, 0, 1)
}
