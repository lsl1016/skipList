package skiplist

import "math/rand/v2"

type skiplist struct {
	head *node
}
type node struct {
	k, v int
	next []*node
}

func newSkiplist(level int) *skiplist {
	return &skiplist{
		head: &node{
			k:    0,
			v:    0,
			next: make([]*node, level),
		},
	}
}
func (s *skiplist) Get(k int) (int, bool) {
	if _node := s.search(k); _node != nil {
		return _node.v, true
	}
	return -1, false
}
func (s *skiplist) search(k int) *node {
	_node := s.head

	for i := len(_node.next) - 1; i >= 0; i-- {
		for _node.next[i] != nil && _node.next[i].k < k {
			_node = _node.next[i]
		}
		if _node.next[0] != nil && _node.next[0].k == k {
			return _node.next[0]
		}
	}
	return nil
}

// 将k，v对插入到跳表中
func (s *skiplist) Put(k, v int) {
	if _node := s.search(k); _node != nil {
		_node.v = v
		return
	}
	level := s.randomLevel()
	// 对高度进行补齐
	for len(s.head.next) < level {
		s.head.next = append(s.head.next, nil)
	}
	newNode := &node{
		k:    k,
		v:    v,
		next: make([]*node, level),
	}
	move := s.head
	for i := level; i > 0; i-- {
		for move.next[i] != nil && move.next[i].k < k {
			move = move.next[i]
		}
		newNode.next[i] = move.next[i]
		move.next[i] = newNode
	}
}
func (s *skiplist) randomLevel() int {
	level := 1
	for rand.Float64() < 0.5 {
		level++
	}
	return level
}
func (s *skiplist) Delete(k int) {
	if _node := s.search(k); _node == nil {
		return
	}
	move := s.head
	for i := len(move.next) - 1; i > 0; i-- {
		for move.next[i] != nil && move.next[i].k < k {
			move = move.next[i]
		}
		if move.next[i] != nil && move.next[i].k == k {
			move.next[i] = move.next[i].next[i]
		}
	}
	var dif int
	for i := len(s.head.next); i > 0 && s.head.next[i] == nil; i-- {
		dif++
	}
	if dif > 0 {
		s.head.next = s.head.next[:len(s.head.next)-dif]
	}
}
// 找到skipList 当中 >= l 且 <=r 的 k v 对
func (s *skiplist) Range(l, r int) [][2]int {

	ceilNode:=s.ceiling(l)
	if ceilNode==nil{
		return [][2]int{}
	}
	var res [][2]int
	for move:=ceilNode;move!=nil && move.k<=r;move=move.next[0]{
		res=append(res,[2]int{move.k,move.v})
	}
	return res
}

// 返回 key 值大于等于 k 且最接近 k 的节点
func (s *skiplist) ceiling(k int) *node {
	_node := s.head
	for i := len(_node.next) - 1; i >= 0; i-- {
		for _node.next[i] != nil && _node.next[i].k < k {
			_node = _node.next[i]
		}
		if _node.next[i] != nil && _node.next[i].k == k {
			return _node.next[i]
		}
	}
	return _node.next[0]
}

func (s *skiplist) Ceiling (k int) ([2]int,bool){
	if _node:=s.ceiling(k);_node!=nil{
		return [2]int{_node.k,_node.v},true
	}
	return [2]int{-1,-1}, false
}

func (s *skiplist) floor(k int) *node {
	_node := s.head
	for i := len(_node.next) - 1; i >= 0; i-- {
		for _node.next[i] != nil && _node.next[i].k < k {
			_node = _node.next[i]
		}
		if _node.next[i] != nil && _node.next[i].k == k {
			return _node.next[i]
		}
	}
	return _node
}
func (s *skiplist) Floor(k int) ([2]int, bool) {
	if _node := s.floor(k); _node != nil {
		return [2]int{_node.k, _node.v}, true
	}
	return [2]int{-1, -1}, false
}
