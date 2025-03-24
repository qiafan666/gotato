package gconsistent

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

var ErrEmptyCircle = errors.New("empty circle")

// Consistent 实现一致性哈希算法的结构体
type Consistent struct {
	circle           map[uint32]string // 哈希环，使用哈希值映射到节点
	sortedHashes     []uint32          // 排序后的哈希值列表，用于快速查找
	NumberOfReplicas int               // 每个节点的虚拟节点数量
	sync.RWMutex                       // 读写锁，保证并发安全
}

// New 创建一个 Consistent 的新实例
// 用于无状态的节点分配，具体可看测试用例
// replicasCount 表示每个节点的虚拟节点数量
func New(elts ...string) *Consistent {
	if len(elts) == 0 {
		panic("empty circle")
	}
	c := new(Consistent)
	c.NumberOfReplicas = len(elts)
	c.circle = make(map[uint32]string)

	for _, elt := range elts {
		c.Add(elt)
	}
	return c
}

// eltKey 为某节点生成虚拟节点的唯一键
// elt 是节点标识，idx 是虚拟节点的编号
func (c *Consistent) eltKey(elt string, idx int) string {
	return strconv.Itoa(idx) + elt
}

// Add 添加一个新节点到哈希环
func (c *Consistent) Add(elt string) {
	c.Lock()
	defer c.Unlock()
	c.add(elt)
}

// add 添加节点并创建其对应的虚拟节点到哈希环中
func (c *Consistent) add(elt string) {
	for i := 0; i < c.NumberOfReplicas; i++ {
		c.circle[c.hashKey(c.eltKey(elt, i))] = elt
	}
	c.updateSortedHashes() // 更新排序的哈希值列表
}

// Remove 从哈希环中移除指定节点
func (c *Consistent) Remove(elt string) {
	c.Lock()
	defer c.Unlock()
	c.remove(elt)
}

// remove 从哈希环中移除指定节点的所有虚拟节点
func (c *Consistent) remove(elt string) {
	for i := 0; i < c.NumberOfReplicas; i++ {
		delete(c.circle, c.hashKey(c.eltKey(elt, i)))
	}
	c.updateSortedHashes() // 更新排序的哈希值列表
}

// Get 根据键获取哈希环上的最近节点
// name 是要查找的键，返回映射到的节点名称
func (c *Consistent) Get(name string) (string, error) {
	c.RLock()
	defer c.RUnlock()
	if len(c.circle) == 0 {
		return "", ErrEmptyCircle
	}
	key := c.hashKey(name)
	i := c.search(key) // 在排序的哈希值中查找
	return c.circle[c.sortedHashes[i]], nil
}

// search 在排序后的哈希列表中查找大于给定键的第一个哈希值的索引
func (c *Consistent) search(key uint32) (i int) {
	f := func(x int) bool {
		return c.sortedHashes[x] > key
	}
	i = sort.Search(len(c.sortedHashes), f)
	if i >= len(c.sortedHashes) {
		i = 0 // 如果超过最大值，返回第一个元素，形成环状结构
	}
	return
}

// hashKey 将给定的键转换为 uint32 哈希值
func (c *Consistent) hashKey(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)]) // 计算 CRC32 哈希
	}
	return crc32.ChecksumIEEE([]byte(key))
}

// updateSortedHashes 更新排序的哈希值列表，确保哈希环的顺序
func (c *Consistent) updateSortedHashes() {
	var hashes []uint32
	for k := range c.circle {
		hashes = append(hashes, k)
	}

	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})

	c.sortedHashes = hashes
}
