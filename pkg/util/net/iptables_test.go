package net

import (
	"context"
	"testing"
)

func TestIPTables(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	iptables := IPTables{}
	ctx := context.Background()

	// should apt-get install iptables manually.
	if err := iptables.Drop(ctx, "n1"); err != nil {
		t.Fatalf("drop network failed %v", err)
	}

	if err := iptables.Heal(ctx); err != nil {
		t.Fatalf("heal netwrok failed %v", err)
	}

	if err := iptables.Fast(ctx); err != nil {
		t.Fatalf("fast netwrok failed %v", err)
	}

	if err := iptables.Slow(ctx, DefaultSlowOptions()); err != nil {
		t.Fatalf("slow netwrok failed %v", err)
	}

	if err := iptables.Fast(ctx); err != nil {
		t.Fatalf("fast netwrok failed %v", err)
	}

	if err := iptables.Flaky(ctx); err != nil {
		t.Fatalf("flaky netwrok failed %v", err)
	}

	if err := iptables.Fast(ctx); err != nil {
		t.Fatalf("fast netwrok failed %v", err)
	}
}
