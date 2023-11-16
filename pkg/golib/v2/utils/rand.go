package utils

import (
	"math/rand"
	"sync"
)

// math/rand.lockedSource 的重新实现
// 解决random包中默认初始化的globalRand的seed只能为1不能自定义的问题
type LockedSource struct {
	mut sync.Mutex
	src rand.Source
}

// NewRand returns a rand.Rand that is threadsafe.
func NewRand(seed int64) *rand.Rand {
	return rand.New(&LockedSource{src: rand.NewSource(seed)})
}

func (r *LockedSource) Int63() (n int64) {
	r.mut.Lock()
	n = r.src.Int63()
	r.mut.Unlock()
	return
}

// Seed implements Seed() of Source
func (r *LockedSource) Seed(seed int64) {
	r.mut.Lock()
	r.src.Seed(seed)
	r.mut.Unlock()
}
