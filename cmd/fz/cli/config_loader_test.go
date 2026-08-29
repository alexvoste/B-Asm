package cli

import (
	"testing"

	"github.com/forgezero-cli/ForgeZero/internal/config"
)

func TestApplyConfigToFlagsUsesConfiguredWorkers(t *testing.T) {
	flags := &Flags{}
	cfg := &config.Config{}
	cfg.Concurrency.Workers = 3

	ApplyConfigToFlags(cfg, flags)

	if flags.Jobs != 3 {
		t.Fatalf("expected configured workers to set jobs, got %d", flags.Jobs)
	}
}

func TestApplyConfigToFlagsPreservesExplicitJobs(t *testing.T) {
	flags := &Flags{Jobs: 7}
	cfg := &config.Config{}
	cfg.Concurrency.Workers = 3

	ApplyConfigToFlags(cfg, flags)

	if flags.Jobs != 7 {
		t.Fatalf("expected explicit jobs to win, got %d", flags.Jobs)
	}
}
