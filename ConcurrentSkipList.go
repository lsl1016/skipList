package skiplist

import (
	"sync"
	"sync/atomic"
)

type ConcurrentSkipList struct {
	// 当前跳表中存在的元素
	cap atomic.Int32

	DeleteMutex sync.RWMutex
	keytoMutex  sync.Map
	head        *node
	nodesCache  sync.Pool
	compareFunc func(k1, k2 any) bool
}
