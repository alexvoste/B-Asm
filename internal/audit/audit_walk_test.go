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

package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanVendorWalkPermission(t *testing.T) {
	root := t.TempDir()
	vendor := filepath.Join(root, "vendor")
	if err := os.MkdirAll(vendor, 0o755); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(vendor, "blocked")
	if err := os.Mkdir(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(blocked, 0o755) }()
	findings := []Finding{}
	seen := map[string]bool{}
	err := scanVendor(context.Background(), root, vendor, nil, &findings, seen)
	if err == nil {
		t.Fatal("expected walk error")
	}
}

func TestScanSecretsWalkError(t *testing.T) {
	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(blocked, 0o755) }()
	findings := []Finding{}
	seen := map[string]bool{}
	err := scanSecrets(context.Background(), root, nil, &findings, seen)
	if err == nil {
		t.Fatal("expected walk error")
	}
}

func TestScanVendorLicensesMissing(t *testing.T) {
	root := t.TempDir()
	findings := []Finding{}
	seen := map[string]bool{}
	if err := scanVendorLicenses(context.Background(), filepath.Join(root, "missing"), nil, &findings, seen); err != nil {
		t.Fatal(err)
	}
}
