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

func TestHashCacheRoundTrip(t *testing.T) {
	cacheDir := t.TempDir()
	want := map[string]hashCacheEntry{
		"src/a.c": {hash: [32]byte{1, 2, 3}, size: 42, modTime: 123456789},
		"src/b.c": {hash: [32]byte{9, 8, 7}, size: 84, modTime: 987654321},
	}
	if err := saveHashCache(cacheDir, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadHashCache(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("loaded %d entries, want %d", len(got), len(want))
	}
	for path, expected := range want {
		if got[path] != expected {
			t.Fatalf("entry %q = %#v, want %#v", path, got[path], expected)
		}
	}
}
