package gid

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"
)

/*
 * int64 ID生成器
 * 编码规则:
 * 32位时间戳 + 12位ServerId + 20位自增ID
 * 支持每秒生成 2^20 > 104万 个ID
 * 每个节点需要用集群唯一的 serverID 来初始化该 ID 生成器
 */

type IDGenerator struct {
	// serverID 服务器ID
	serverID int64
	// idCounter ID计数器
	idCounter int64
	// 以下全局变量仅被NewID()使用，并且被mutex保证互斥访问
	// 互斥访问 idCounter 和 lastCounterResetTs
	mutex *sync.Mutex
	// 上一次Counter重置时的时间戳
	lastCounterResetTs int64
}

// CounterBits ID编码中自增器所占位数
const CounterBits = 20

// MaxCounterInSecond 单秒允许生成的最大ID数量
const MaxCounterInSecond = (1 << CounterBits) - 1

// serverIDList 服务器ID列表
var serverIDList []int32

// NewIDGenerator 初始化服务器ID
func NewIDGenerator(sid int32) *IDGenerator {

	if sid > 4095 {
		panic(fmt.Sprintf("sid %d is out of range", sid))
	}

	serverIDList = append(serverIDList, sid)
	return &IDGenerator{
		serverID:           int64(sid),
		idCounter:          0,
		mutex:              &sync.Mutex{},
		lastCounterResetTs: 0,
	}
}

// NewID 获取一个唯一ID
func (idGen *IDGenerator) NewID() int64 {
	now := time.Now().Unix()
	idGen.mutex.Lock()
	defer idGen.mutex.Unlock()
	if now != idGen.lastCounterResetTs {
		idGen.idCounter = 0
		idGen.lastCounterResetTs = now
	} else {
		idGen.idCounter++
		// 单秒生成的ID超过最大限制，视为严重Bug
		// 1. 线上环境，DPanic不会让程序直接panic，避免不严重的overflow引起严重的问题
		// 2. 发生overflow，idCounter=MaxCounterInSecond 继续运行，便于后续找到出问题的ID
		if idGen.idCounter > MaxCounterInSecond {
			idGen.idCounter = MaxCounterInSecond
			log.Panicf("idgen: id counter overflow, serverID: %d, idCounter: %d", idGen.serverID, idGen.idCounter)
		}
	}
	return (now << 32) | (idGen.serverID << CounterBits) | idGen.idCounter
}

// ParseID 解析一个id
func ParseID(id int64) (serverID int32, createAt int64) {
	createAt = (id >> 32) * 1000
	serverID = int32(uint32(id) >> CounterBits)
	return serverID, createAt
}

func IsServerIDValid(id int32) bool {
	return slices.Contains(serverIDList, id)
}
