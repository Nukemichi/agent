package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agent-michi/internal/domain"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const cacheTTL = 2 * time.Second

// Collector implements domain.StatsCollector with an in-memory TTL cache.
type Collector struct {
	mu        sync.Mutex
	cached    domain.StatsResponse
	cachedAt  time.Time
}

// NewCollector creates a new Collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Collect returns cached stats if within TTL, otherwise gathers fresh metrics.
func (c *Collector) Collect(ctx context.Context) (domain.StatsResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.cachedAt) < cacheTTL {
		return c.cached, nil
	}

	stats, err := gather(ctx)
	if err != nil {
		return domain.StatsResponse{}, err
	}

	c.cached = stats
	c.cachedAt = time.Now()
	return stats, nil
}

func gather(_ context.Context) (domain.StatsResponse, error) {
	cpuPcts, err := cpu.Percent(0, false)
	if err != nil {
		return domain.StatsResponse{}, fmt.Errorf("collect cpu: %w", err)
	}
	var cpuPct float64
	if len(cpuPcts) > 0 {
		cpuPct = cpuPcts[0]
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return domain.StatsResponse{}, fmt.Errorf("collect memory: %w", err)
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return domain.StatsResponse{}, fmt.Errorf("collect disk: %w", err)
	}

	hostStat, err := host.Info()
	if err != nil {
		return domain.StatsResponse{}, fmt.Errorf("collect host: %w", err)
	}

	return domain.StatsResponse{
		CPU:       cpuPct,
		RAMUsed:   vmStat.Used / 1024 / 1024,
		RAMTotal:  vmStat.Total / 1024 / 1024,
		DiskUsed:  diskStat.Used / 1024 / 1024 / 1024,
		DiskTotal: diskStat.Total / 1024 / 1024 / 1024,
		Uptime:    hostStat.Uptime,
	}, nil
}
