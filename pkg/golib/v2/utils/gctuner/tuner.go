package gctuner

import (
	"os"
	"runtime/debug"
	"strconv"
	"sync/atomic"
)

var (
	maxGCPercent     uint32 = 500
	minGCPercent     uint32 = 50
	defaultGCPercent uint32 = 100
)

/*
userPercent ∈ [0,100] 表示达到总内存的 userPercent/100 时触发gc
userPercent <=0 , 表示关闭自适应调节功能
userPercent >=100 , 表示内存使用无上限，可能会OOM
*/
func Tuning(userPercent uint32) {
	// disable gc tuner if percent <= 0
	if userPercent <= 0 {
		if globalTuner != nil {
			globalTuner.stop()
			globalTuner = nil
		}
		return
	}

	if globalTuner == nil {
		globalTuner = newTuner(userPercent)
		return
	}
	globalTuner.setThreshold(userPercent)
}

// GetGCPercent returns the current GCPercent.
func GetGCPercent() uint32 {
	if globalTuner == nil {
		return defaultGCPercent
	}
	return globalTuner.getGCPercent()
}

func newTuner(userPercent uint32) *tuner {
	gogcEnv := os.Getenv("GOGC")
	gogc, err := strconv.ParseInt(gogcEnv, 10, 32)
	if err == nil {
		// 默认的 gcPercent 还是优先从 GOGC 环境变量中读取，读不到使用默认值 100
		defaultGCPercent = uint32(gogc)
	}

	t := &tuner{
		gcPercent:   defaultGCPercent,
		userPercent: userPercent,
	}
	t.finalizer = newFinalizer(t.tuning) // start tuning

	return t
}

// only allow one gc tuner in one process
var globalTuner *tuner = nil

type tuner struct {
	finalizer   *finalizer
	gcPercent   uint32 // stored
	userPercent uint32 // percentage of total memory, ex=80 => 80% * totalRAM
}

// tuning check the memory inuse and tune GC percent dynamically.
// Go runtime ensure that it will be called serially.
func (t *tuner) tuning() {
	userPercent := t.getUserPercent()

	// stop gc tuning
	if userPercent <= 0 || userPercent > 100 {
		return
	}

	used, err := MemUsed()
	if err != nil {
		return
	}
	total, err := MemTotal()
	if err != nil {
		return
	}

	// RSS / total
	memPercent := float64(used) * 100 / float64(total)

	t.setGCPercent(calcGCPercent(userPercent, memPercent))
	return
}

/* Heap
 _______________  => limit: host/cgroup memory hard limit
|               |
|---------------| => threshold: increase GCPercent when gc_trigger < threshold
|               |
|---------------| => gc_trigger: heap_live + heap_live * GCPercent / 100
|               |
|---------------|
|   heap_live   |
|_______________|
*/
func calcGCPercent(userPercent uint32, memPercent float64) uint32 {
	// memPercent = (inuse / total) * 100
	// gcPercent := uint32(math.Floor(float64(threshold-inuse) / float64(inuse) * 100))

	gcPercent := uint32((float64(userPercent) - memPercent) / memPercent * 100)

	if gcPercent < minGCPercent {
		return minGCPercent
	} else if gcPercent > maxGCPercent {
		return maxGCPercent
	}

	return gcPercent
}

func (t *tuner) stop() {
	t.finalizer.stop()
}

func (t *tuner) setThreshold(userPercent uint32) {
	atomic.StoreUint32(&t.userPercent, userPercent)
}

func (t *tuner) getUserPercent() uint32 {
	return atomic.LoadUint32(&t.userPercent)
}

func (t *tuner) setGCPercent(percent uint32) uint32 {
	if t.gcPercent == percent {
		return t.gcPercent
	}

	atomic.StoreUint32(&t.gcPercent, percent)
	return uint32(debug.SetGCPercent(int(percent)))
}

func (t *tuner) getGCPercent() uint32 {
	return atomic.LoadUint32(&t.gcPercent)
}

// GetMaxGCPercent returns the max gc percent value.
func GetMaxGCPercent() uint32 {
	return atomic.LoadUint32(&maxGCPercent)
}

// SetMaxGCPercent sets the new max gc percent value.
func SetMaxGCPercent(n uint32) uint32 {
	return atomic.SwapUint32(&maxGCPercent, n)
}

// GetMinGCPercent returns the min gc percent value.
func GetMinGCPercent() uint32 {
	return atomic.LoadUint32(&minGCPercent)
}

// SetMinGCPercent sets the new min gc percent value.
func SetMinGCPercent(n uint32) uint32 {
	return atomic.SwapUint32(&minGCPercent, n)
}
