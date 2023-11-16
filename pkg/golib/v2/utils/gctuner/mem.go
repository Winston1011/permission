package gctuner

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

const (
	cGroupMemLimitPath = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	cGroupMemUsagePath = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	selfCGroupPath     = "/proc/self/cgroup"
)

// MemTotal returns the total amount of RAM on this system
var MemTotal func() (uint64, error)

// MemUsed returns the total used amount of RAM on this system
var MemUsed func() (uint64, error)

// expiration time is 60min
var memLimit *memInfoCache

// expiration time is 500ms
var memUsage *memInfoCache

func init() {
	if inContainer() {
		MemTotal = memTotalCGroup
		MemUsed = memUsedCGroup
	} else {
		MemTotal = memTotalNormal
		MemUsed = memUsedNormal
	}
	memLimit = &memInfoCache{
		RWMutex: &sync.RWMutex{},
	}
	memUsage = &memInfoCache{
		RWMutex: &sync.RWMutex{},
	}
}

// MemTotalCGroup returns the total amount of RAM on this system in container environment.
func memTotalCGroup() (uint64, error) {
	memo, t := memLimit.get()
	if time.Since(t) < 60*time.Minute {
		return memo, nil
	}
	memo, err := readUint(cGroupMemLimitPath)
	if err != nil {
		return memo, err
	}
	memLimit.set(memo, time.Now())
	return memo, nil
}

// MemUsedCGroup returns the total used amount of RAM on this system in container environment.
func memUsedCGroup() (uint64, error) {
	memo, t := memUsage.get()
	if time.Since(t) < 500*time.Millisecond {
		return memo, nil
	}
	memo, err := readUint(cGroupMemUsagePath)
	if err != nil {
		return memo, err
	}
	memUsage.set(memo, time.Now())
	return memo, nil
}

// MemTotalNormal returns the total amount of RAM on this system in non-container environment.
func memTotalNormal() (uint64, error) {
	total, t := memLimit.get()
	if time.Since(t) < 60*time.Minute {
		return total, nil
	}
	v, err := mem.VirtualMemory()
	if err != nil {
		return v.Total, err
	}
	memLimit.set(v.Total, time.Now())
	return v.Total, nil
}

// MemUsedNormal returns the total used amount of RAM on this system in non-container environment.
func memUsedNormal() (uint64, error) {
	used, t := memUsage.get()
	if time.Since(t) < 500*time.Millisecond {
		return used, nil
	}
	v, err := mem.VirtualMemory()
	if err != nil {
		return v.Used, err
	}
	memUsage.set(v.Used, time.Now())
	return v.Used, nil
}

func inContainer() bool {
	v, err := os.ReadFile(selfCGroupPath)
	if err != nil {
		return false
	}
	if strings.Contains(string(v), "docker") ||
		strings.Contains(string(v), "kubepods") ||
		strings.Contains(string(v), "containerd") {
		return true
	}
	return false
}

// refer to https://github.com/containerd/cgroups/blob/318312a373405e5e91134d8063d04d59768a1bff/utils.go#L251
func parseUint(s string, base, bitSize int) (uint64, error) {
	v, err := strconv.ParseUint(s, base, bitSize)
	if err != nil {
		intValue, intErr := strconv.ParseInt(s, base, bitSize)
		// 1. Handle negative values greater than MinInt64 (and)
		// 2. Handle negative values lesser than MinInt64
		if intErr == nil && intValue < 0 {
			return 0, nil
		} else if intErr != nil &&
			intErr.(*strconv.NumError).Err == strconv.ErrRange &&
			intValue < 0 {
			return 0, nil
		}
		return 0, err
	}
	return v, nil
}

// refer to https://github.com/containerd/cgroups/blob/318312a373405e5e91134d8063d04d59768a1bff/utils.go#L243
func readUint(path string) (uint64, error) {
	v, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return parseUint(strings.TrimSpace(string(v)), 10, 64)
}

type memInfoCache struct {
	*sync.RWMutex
	mem        uint64
	updateTime time.Time
}

func (c *memInfoCache) get() (memo uint64, t time.Time) {
	c.RLock()
	defer c.RUnlock()
	memo, t = c.mem, c.updateTime
	return
}

func (c *memInfoCache) set(memo uint64, t time.Time) {
	c.Lock()
	defer c.Unlock()
	c.mem, c.updateTime = memo, t
}
