package consistenthash

import (
	"container/heap"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// 实现堆的接口
type Keys []int

func (keys Keys) Len() int {
	return len(keys)
}

func (keys Keys) Less(i, j int) bool {
	return keys[i] < keys[j]
}

func (keys Keys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

func (keys *Keys) Push(x interface{}) {
	*keys = append(*keys, x.(int))
}

func (keys *Keys) Pop() interface{} {
	res := (*keys)[len(*keys)-1]
	*keys = (*keys)[:len(*keys)-1]
	return res
}

// Map constains all hashed keys
// 利用一致性哈希算法存储peer节点，解决少量增减服务器导致的大量震荡问题
type Map struct {
	hash Hash
	//每个节点拥有的虚拟节点个数
	replicas int
	// 哈希环， 节点的哈希值
	keys Keys
	// 哈希值对应的节点
	hashMap map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 默认hash函数
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	heap.Init(&m.keys)
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			heap.Push(&m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
