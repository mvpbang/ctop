package collector

import (
	"testing"

	api "github.com/fsouza/go-dockerclient"
)

func TestReadMemCgroupV2(t *testing.T) {
	d := &Docker{}
	stats := &api.Stats{}
	stats.MemoryStats.Usage = 1000 * 1024 * 1024 // 1000 MB
	stats.MemoryStats.Limit = 2000 * 1024 * 1024 // 2000 MB
	// 在 cgroups v2 下，Cache 为 0，而 InactiveFile 为 200MB
	stats.MemoryStats.Stats.Cache = 0
	stats.MemoryStats.Stats.InactiveFile = 200 * 1024 * 1024

	d.ReadMem(stats)

	expectedUsage := int64(800 * 1024 * 1024)
	if d.MemUsage != expectedUsage {
		t.Fatalf("expected usage %d, got %d", expectedUsage, d.MemUsage)
	}
	expectedPercent := 40
	if d.MemPercent != expectedPercent {
		t.Fatalf("expected percent %d, got %d", expectedPercent, d.MemPercent)
	}
}
