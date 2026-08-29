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

package linker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/forgezero-cli/ForgeZero/internal/utils"
	"github.com/forgezero-cli/ForgeZero/internal/zig"
)

func TestCreateResponseFileOK(t *testing.T) {
	path, err := createResponseFile([]string{"a.o", "-o", "bin"})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "a.o") {
		t.Fatal(string(data))
	}
}

func TestCreateResponseFileInvalidArg(t *testing.T) {
	_, err := createResponseFile([]string{"bad\narg"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateResponseFileValidateFail(t *testing.T) {
	_, err := createResponseFile([]string{"bad;arg"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShouldUseResponseFileCoverage(t *testing.T) {
	args := make([]string, 200)
	for i := range args {
		args[i] = "x"
	}
	if !shouldUseResponseFile(args) {
		t.Fatal("expected response file")
	}
	long := strings.Repeat("a", 9000)
	if !shouldUseResponseFile([]string{long}) {
		t.Fatal("expected length trigger")
	}
}

func TestLdForTargetBranches(t *testing.T) {
	old := Target
	defer func() { Target = old }()
	cases := []struct {
		target string
		want   string
	}{
		{"x86_64-linux-gnu", "ld"},
		{"arm-linux-gnueabihf", "arm-linux-gnueabihf-ld"},
		{"riscv64-unknown-elf", "riscv64-unknown-elf-ld"},
		{"wasm32-unknown-unknown", "wasm-ld"},
	}
	for _, tc := range cases {
		Target = tc.target
		got := ldForTarget()
		if tc.target == "wasm32-unknown-unknown" {
			if got != "wasm-ld" && got != "ld.lld" {
				t.Fatalf("%s: %s", tc.target, got)
			}
			continue
		}
		if got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.target, got, tc.want)
		}
	}
}

func TestGccForTargetBranches(t *testing.T) {
	old := Target
	defer func() { Target = old }()
	Target = "arm-linux-gnueabihf"
	if gccForTarget() != "arm-linux-gnueabihf-gcc" {
		t.Fatal(gccForTarget())
	}
	Target = "wasm32-unknown-unknown"
	g := gccForTarget()
	if g != "emcc" && g != "clang" {
		t.Fatal(g)
	}
}

func TestLinkWithZigMock(t *testing.T) {
	oldRunner := runner
	oldZigReq := zig.ZigRequested
	oldRun := zig.RunCommand
	oldCheck := utils.CheckToolFunc
	defer func() {
		runner = oldRunner
		zig.ZigRequested = oldZigReq
		zig.RunCommand = oldRun
		utils.CheckToolFunc = oldCheck
	}()
	utils.CheckToolFunc = func(string) error { return nil }
	zig.ZigRequested = true
	zig.RunCommand = func(ctx context.Context, verbose bool, args ...string) (string, error) {
		return "", nil
	}
	err := linkWithZig(context.Background(), []string{"a.o"}, "out", false, "x86_64-linux-gnu", false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLinkWithZigUnavailable(t *testing.T) {
	oldCheck := utils.CheckToolFunc
	defer func() { utils.CheckToolFunc = oldCheck }()
	utils.CheckToolFunc = func(string) error { return errors.New("no zig") }
	zig.ZigRequested = true
	defer func() { zig.ZigRequested = false }()
	err := linkWithZig(context.Background(), []string{"a.o"}, "out", false, "", false, false, nil)
	if err == nil || !strings.Contains(err.Error(), "zig") {
		t.Fatalf("got %v", err)
	}
}

func TestLinkWithGccFallbackDetachedNoPie(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	callCount := 0
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "fail", errors.New("fail")
		}
		return "ok", nil
	}}
	if err := linkWithGcc(context.Background(), []string{"a.o"}, "out", false, true, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Fatalf("expected fallback attempt, got %d", callCount)
	}
}

func TestLinkWithGccFallbackNoPieVerboseReporting(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	callCount := 0
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		callCount++
		if callCount == 1 {
			return "first fail", errors.New("fail")
		}
		return "ok", nil
	}}
	if err := linkWithGcc(context.Background(), []string{"a.o"}, "out", true, true, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Fatalf("expected fallback attempt, got %d", callCount)
	}
}

func TestLinkWithLdMockCoverage(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "", nil
	}}
	err := linkWithLd(context.Background(), []string{"a.o", "b.o"}, "out", false, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunLinkerCommandResponseFile(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	var gotArgs []string
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		gotArgs = args
		return "", nil
	}}
	big := make([]string, 150)
	for i := range big {
		big[i] = "arg"
	}
	_, err := runLinkerCommand(context.Background(), false, "ld", big)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotArgs) != 1 || !strings.HasPrefix(gotArgs[0], "@") {
		t.Fatalf("got %v", gotArgs)
	}
}

func TestSplitObjectsPartitioning(t *testing.T) {
	parts := splitObjects([]string{"a.o", "b.o", "c.o", "d.o", "e.o"}, 3)
	if len(parts) != 3 {
		t.Fatalf("expected 3 partitions, got %d", len(parts))
	}
	if len(parts[0]) != 2 || len(parts[1]) != 2 || len(parts[2]) != 1 {
		t.Fatalf("unexpected partition sizes %v", []int{len(parts[0]), len(parts[1]), len(parts[2])})
	}
}

func TestLinkPartitionAndFinalGetErrors(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "partial fail", errors.New("fail")
	}}
	if err := linkPartition(context.Background(), &LinkTarget{Name: "tmp.o", Objs: []string{"a.o"}}, "raw", false); err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkFinalReturnsVerboseError(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "final fail", errors.New("fail")
	}}
	if err := linkFinal(context.Background(), &LinkTarget{Name: "out", Objs: []string{"a.o"}}, "raw", true, false, false, nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestUseZigFlags(t *testing.T) {
	old := ZigEnabled
	ZigEnabled = true
	defer func() { ZigEnabled = old }()
	if !useZig() {
		t.Fatal()
	}
	ZigRequested = true
	defer func() { ZigRequested = false }()
	if !useZig() {
		t.Fatal()
	}
}

func TestDetectLinkerRespectsPreferredLinker(t *testing.T) {
	old := PreferredLinker
	oldOnce := linkerOnce
	oldPreferred := preferredLinker
	oldHasLld := hasLld
	oldHasMold := hasMold
	oldCache := toolPathCache
	PreferredLinker = "mold"
	linkerOnce = &sync.Once{}
	preferredLinker = ""
	hasLld = false
	hasMold = false
	toolPathCache = &sync.Map{}
	defer func() {
		PreferredLinker = old
		linkerOnce = oldOnce
		preferredLinker = oldPreferred
		hasLld = oldHasLld
		hasMold = oldHasMold
		toolPathCache = oldCache
	}()
	detectLinker()
	if preferredLinker != "mold" {
		t.Fatalf("expected preferredLinker mold, got %s", preferredLinker)
	}
	if !hasMold {
		t.Fatal("expected hasMold true")
	}
}

func TestTryAutoLinkStrict(t *testing.T) {
	oldRunner := runner
	oldTarget := Target
	oldCheck := utils.CheckToolFunc
	defer func() {
		runner = oldRunner
		Target = oldTarget
		utils.CheckToolFunc = oldCheck
	}()
	Target = "x86_64-linux-gnu"
	utils.CheckToolFunc = func(name string) error { return nil }
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "", nil
	}}
	dir := t.TempDir()
	obj := filepath.Join(dir, "a.o")
	if err := os.WriteFile(obj, []byte{0x7f, 'E', 'L', 'F', 1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "out")
	err := tryAutoLink(context.Background(), []string{obj}, bin, true, false, true, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReadSymbolsWithObjdumpIntegration(t *testing.T) {
	if _, err := exec.LookPath("objdump"); err != nil {
		t.Skip("objdump not installed")
	}
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "a.c")
	if err := os.WriteFile(src, []byte("int f(void){return 0;}"), 0o644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "a.o")
	if out, err := exec.Command("gcc", "-c", src, "-o", obj).CombinedOutput(); err != nil {
		t.Skip(string(out))
	}
	syms, err := readSymbolsWithObjdump(context.Background(), obj, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) == 0 {
		t.Fatal("expected symbols")
	}
}

func TestReadSymbolsWithReadelfIntegration(t *testing.T) {
	if _, err := exec.LookPath("readelf"); err != nil {
		t.Skip("readelf not installed")
	}
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "b.c")
	if err := os.WriteFile(src, []byte("int g(void){return 1;}"), 0o644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "b.o")
	if out, err := exec.Command("gcc", "-c", src, "-o", obj).CombinedOutput(); err != nil {
		t.Skip(string(out))
	}
	syms, err := readSymbolsWithReadelf(context.Background(), obj, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) == 0 {
		t.Fatal("expected symbols")
	}
}

func TestLinkWithLdVerboseFail(t *testing.T) {
	oldRunner := runner
	defer func() { runner = oldRunner }()
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "ld err", errors.New("fail")
	}}
	err := linkWithLd(context.Background(), []string{"a.o"}, "out", true, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateLinkCallErrorsCoverage(t *testing.T) {
	if err := validateLinkCall(nil, "out"); err == nil { //nolint:staticcheck
		t.Fatal("expected nil ctx error")
	}
	if err := validateLinkCall(context.Background(), ""); err == nil {
		t.Fatal("expected empty output error")
	}
}

func TestLinkWindowsImplMock(t *testing.T) {
	oldRunner := runner
	oldCheck := utils.CheckToolFunc
	defer func() {
		runner = oldRunner
		utils.CheckToolFunc = oldCheck
	}()
	utils.CheckToolFunc = func(string) error { return nil }
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "", nil
	}}
	dir := t.TempDir()
	obj := filepath.Join(dir, "a.o")
	if err := os.WriteFile(obj, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := linkWindowsImpl(context.Background(), []string{obj}, filepath.Join(dir, "out.exe"), false, false, nil); err != nil {
		t.Fatal(err)
	}
}

func TestLinkWindowsImplMultipleMock(t *testing.T) {
	oldRunner := runner
	oldCheck := utils.CheckToolFunc
	defer func() {
		runner = oldRunner
		utils.CheckToolFunc = oldCheck
	}()
	utils.CheckToolFunc = func(string) error { return nil }
	runner = &MockRunner{RunFunc: func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		return "", nil
	}}
	dir := t.TempDir()
	obj := filepath.Join(dir, "a.o")
	if err := os.WriteFile(obj, []byte{1, 2, 3}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := linkWindowsImpl(context.Background(), []string{obj}, filepath.Join(dir, "out.exe"), true, true, []string{"m"}); err != nil {
		t.Fatal(err)
	}
}

func TestReadSymbolsFallbackObjdump(t *testing.T) {
	if _, err := exec.LookPath("objdump"); err != nil {
		t.Skip("objdump required")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "a.c")
	if err := os.WriteFile(src, []byte("int f(void){return 0;}"), 0o644); err != nil {
		t.Fatal(err)
	}
	obj := filepath.Join(dir, "a.o")
	if out, err := exec.Command("gcc", "-c", src, "-o", obj).CombinedOutput(); err != nil {
		t.Skip(string(out))
	}
	syms, err := readSymbolsWithObjdump(context.Background(), obj, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(syms) == 0 {
		t.Fatal("expected symbols")
	}
}
