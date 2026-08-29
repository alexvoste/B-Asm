package assembler

import (
	"context"
	"strings"
	"testing"

	"github.com/forgezero-cli/ForgeZero/internal/config"
	"github.com/forgezero-cli/ForgeZero/internal/utils"
)

func TestCompileCUsesConfiguredCompilerAndFlags(t *testing.T) {
	var gotName string
	var gotArgs []string
	oldRun := runCommand
	defer func() { runCommand = oldRun }()
	SetRunCommand(func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return "", nil
	})

	ctx := utils.ContextWithConfig(context.Background(), &config.Config{
		Compiler: config.CompilerConfig{Path: "/custom/cc"},
		Flags:    config.Flags{Cc: []string{"-O3", "-fno-plt"}},
	})
	if err := compileCWithTarget(ctx, "source.c", "source.o", false, "gcc", "x86_64-linux-gnu"); err != nil {
		t.Fatal(err)
	}

	if gotName != "/custom/cc" {
		t.Fatalf("expected configured compiler, got %q", gotName)
	}
	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "-O3") || !strings.Contains(joined, "-fno-plt") {
		t.Fatalf("configured compiler flags missing from %v", gotArgs)
	}
}

func TestAssembleNasmUsesConfiguredFlags(t *testing.T) {
	var gotArgs []string
	oldRun := runCommand
	defer func() { runCommand = oldRun }()
	SetRunCommand(func(ctx context.Context, verbose bool, name string, args ...string) (string, error) {
		gotArgs = append([]string(nil), args...)
		return "", nil
	})

	ctx := utils.ContextWithConfig(context.Background(), &config.Config{
		Flags: config.Flags{Asm: []string{"-g", "-w+all"}},
	})
	if err := assembleWithNasm(ctx, "source.asm", "source.o", false, false, "x86_64-linux-gnu"); err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(gotArgs, " ")
	if !strings.Contains(joined, "-g") || !strings.Contains(joined, "-w+all") {
		t.Fatalf("configured assembler flags missing from %v", gotArgs)
	}
}
