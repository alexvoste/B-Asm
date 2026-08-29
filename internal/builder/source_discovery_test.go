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

package builder

import (
	"path/filepath"
	"testing"
)

func TestSourceDiscoveryRootsAreUnique(t *testing.T) {
	dir := t.TempDir()
	roots := []string{dir, filepath.Join(dir, "src"), filepath.Join(dir, "src")}
	seen := make(map[string]struct{}, len(roots))
	unique := roots[:0]
	for _, root := range roots {
		key, err := filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		key = filepath.Clean(key)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, root)
	}
	if len(unique) != 2 {
		t.Fatalf("unique roots = %d, want 2", len(unique))
	}
}
