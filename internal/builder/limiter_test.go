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

package builder

import "testing"

func TestParseMemInfo(t *testing.T) {
	data := []byte("MemTotal:       16384 kB\nMemFree:\t2048 kB\nMemAvailable: 8192 kB\n")
	available, free := parseMemInfo(data)
	if available != 8192 || free != 2048 {
		t.Fatalf("parseMemInfo() = (%d, %d), want (8192, 2048)", available, free)
	}
}

func TestParseMemInfoIgnoresMalformedValues(t *testing.T) {
	data := []byte("MemAvailable: nope\nMemFree:\nMemAvailable: 4096 kB\n")
	available, free := parseMemInfo(data)
	if available != 4096 || free != 0 {
		t.Fatalf("parseMemInfo() = (%d, %d), want (4096, 0)", available, free)
	}
}
