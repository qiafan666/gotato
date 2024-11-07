package gconsistent

import (
	"fmt"
	"testing"
)

func TestConsistentHash(t *testing.T) {

	// 创建一个一致性哈希实例，每个节点有3个虚拟节点
	consistent := New(3)

	get, err := consistent.Get("User1")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(get)
	// 添加三个节点到哈希环上
	consistent.Add("NodeA")
	consistent.Add("NodeB")
	consistent.Add("NodeC")

	// 模拟请求，根据请求 key 分配到合适的节点
	requests := []string{"User1", "User2", "User3", "User4"}
	for _, req := range requests {
		node, err := consistent.Get(req)
		if err != nil {
			fmt.Printf("Error retrieving node for request %s: %s\n", req, err)
		} else {
			fmt.Printf("Request %s is mapped to node %s\n", req, node)
		}
	}

	// 如果移除一个节点
	consistent.Remove("NodeB")
	fmt.Println("\nAfter removing NodeB:")
	for _, req := range requests {
		node, err := consistent.Get(req)
		if err != nil {
			fmt.Printf("Error retrieving node for request %s: %s\n", req, err)
		} else {
			fmt.Printf("Request %s is mapped to node %s\n", req, node)
		}
	}
}
