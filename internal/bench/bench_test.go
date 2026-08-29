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

package bench

import (
	"strings"
	"testing"
)

func TestTimerJSON(t *testing.T) {
	timer := NewTimer()
	_ = timer.Stage("check", func() error { return nil })
	data, err := timer.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "total_ns") {
		t.Fatalf("unexpected json: %s", string(data))
	}
}
