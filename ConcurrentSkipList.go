package skiplist

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

type ConcurrentSkipList struct {
	// 当前跳表中存在的元素数量
	cap atomic.Int32

	// 只有删除操作取的是写锁
	DeleteMutex sync.RWMutex

	// 对同一个key 的 put 操作需要互斥
	keytoMutex sync.Map

	// 跳表的头部节点

	head *currentNode

	// 对删除和新建节点进行复用
	nodesCache sync.Pool

	compareFunc func(k1, k2 any) bool
}
type currentNode struct {
	k, v  any
	nexts []*currentNode

	// 用于实现节点的左边界锁
	sync.RWMutex
}

func NewConcurrentSkipList(compareFunc func(kk1, k2 any) bool) *ConcurrentSkipList {
	return &ConcurrentSkipList{
		head: &currentNode{
			nexts: make([]*currentNode, 1),
		},
		nodesCache: sync.Pool{
			New: func() any { return &node{} },
		},
		compareFunc: compareFunc,
	}
}
func (c *ConcurrentSkipList) Delete(k any) {

	c.DeleteMutex.Lock()
	defer c.DeleteMutex.Unlock()

	// 用于接收k 对应的节点，方便复用
	var deleteNode *currentNode

	move := c.head
	for i := len(move.nexts) - 1; i > 0; i-- {
		for move.nexts[i] != nil && c.compareFunc(move.nexts[i].k, k) {
			move = move.nexts[i]
		}
		// 找到了要删除的元素
		if move.nexts[i] != nil && move.nexts[i].k == k {
			deleteNode = move.nexts[i]
			move.nexts[i] = move.nexts[i].nexts[i]
		}
	}
	// 如果 key 不存在，提前返回
	if deleteNode == nil {
		return
	}

	defer c.cap.Add(-1)

	deleteNode.nexts = nil
	c.nodesCache.Put(deleteNode)

	//尝试对跳表的高度进行缩容
	var dif int
	for i := len(c.head.nexts) - 1; i >= 0 && c.head.nexts[i] == nil; i-- {
		dif++
	}
	c.head.nexts = c.head.nexts[:len(c.head.nexts)-dif]
}
func (c *ConcurrentSkipList) Get(k any) (any, bool) {
	c.DeleteMutex.RLock()
	defer c.DeleteMutex.RUnlock()

	node := c.search(k)

	if node != nil {
		return node.v, true
	}
	return nil, false
}
func (c *ConcurrentSkipList) Put(k, v any) {
	c.DeleteMutex.RLock()
	defer c.DeleteMutex.RUnlock()

	// 对同一个key的put操作需要互斥
	keyMutex := c.getKeyToMutex(k)
	keyMutex.Lock()
	defer keyMutex.Unlock()

	// 更新节点
	node := c.search(k)
	if node != nil {
		node.v = v
		return
	}
	// 新增节点
	defer c.cap.Add(1)

	rLevel := c.randomLevel()

	// 通过 sync pool 复用节点
	newNode := c.nodesCache.Get().(*currentNode)
	newNode.k = k
	newNode.v = v
	newNode.nexts = make([]*currentNode, rLevel)

	// 对创建出来的新节点需要加写锁，避免在插入过程中成为 get 流程的左边界
	newNode.Lock()
	defer newNode.Unlock()

	// 如果新节点导致跳表需要扩容，还需要对head 加锁
	if rLevel > len(c.head.nexts)-1 {
		c.head.Lock()
		for i := rLevel; i >= 0; i-- {
			c.head.nexts = append(c.head.nexts, nil)

		}
		c.head.Unlock()
	}
	// 寻找左边界
	move := c.head
	var last *currentNode

	for i := len(c.head.nexts) - 1; i >= 0; i-- {
		for move.nexts[i] != nil && c.compareFunc(move.nexts[i].k, k) {
			move = move.nexts[i]
		}
		if move != last {
			move.RLock()
			defer move.RUnlock()
			last = move
		}

		newNode.nexts[i] = move.nexts[i]
		move.nexts[i] = newNode
	}
}

func (c *ConcurrentSkipList) getKeyToMutex(k any) *sync.Mutex {
	// 基于symc map 来管理 key 锁
	rawMutex, _ := c.keytoMutex.LoadOrStore(k, &sync.Mutex{})
	mutex, _ := rawMutex.(*sync.Mutex)
	return mutex
}
func (c *ConcurrentSkipList) randomLevel() int {
	level := 1
	for rand.Intn(10) < 5 {
		level++
	}
	return level
}

func (c *ConcurrentSkipList) search(k any) *currentNode {
	move := c.head

	// 记录上一层最后节点的位置，避免在下降时对同意节点反复加左边界锁
	var last *currentNode

	for i := len(c.head.nexts) - 1; i >= 0; i-- {
		for move.nexts[i] != nil && c.compareFunc(move.nexts[i].k, k) {
			move = move.nexts[i]
		}
		// 走到左边界
		if move != last {
			move.RLock()
			defer move.RUnlock()
			last = move
		}
		if move.nexts[i] != nil && move.nexts[i].k == k {
			return move.nexts[i]
		}
	}
	return nil
}
