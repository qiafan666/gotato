package gid

import (
	"fmt"
	"log"
	"math/rand"
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

type ID64Generator struct {
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

// serverID64List 服务器ID列表
var serverID64List []int32

// RandID64 随机生成一个服务器ID,返回一个唯一ID
func RandID64() int64 {
	// 使用当前时间作为随机数种子
	rand.Seed(time.Now().UnixNano())
	// 随机生成一个0-4095的整数
	generator := NewID64Generator(int32(rand.Intn(4096)))
	return generator.NewID64()
}

// NewID64Generator 初始化服务器ID
func NewID64Generator(sid int32) *ID64Generator {

	if sid > 4095 {
		panic(fmt.Sprintf("sid %d is out of range", sid))
	}

	serverID64List = append(serverID64List, sid)
	return &ID64Generator{
		serverID:           int64(sid),
		idCounter:          0,
		mutex:              &sync.Mutex{},
		lastCounterResetTs: 0,
	}
}

// NewID64 获取一个唯一ID
func (idGen *ID64Generator) NewID64() int64 {
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

// ParseID64 解析一个id
func ParseID64(id int64) (serverID int32, createAt int64) {
	createAt = (id >> 32) * 1000
	serverID = int32(uint32(id) >> CounterBits)
	return serverID, createAt
}

func IsServerID64Valid(id int32) bool {
	return slices.Contains(serverID64List, id)
}

//////////////////////////////////////////////以下为int32版本//////////////////////////////////////////////

var serverID32List []int32

type ID32Generator struct {
	// serverID 服务器ID
	serverID int32
	// idCounter ID计数器
	idCounter int32
	// 以下全局变量仅被NewID()使用，并且被mutex保证互斥访问
	// 互斥访问 idCounter 和 lastCounterResetTs
	mutex *sync.Mutex
	// 上一次Counter重置时的时间戳
	lastCounterResetTs int64
}

// RandID32 随机生成一个服务器ID，返回一个唯一的int32 ID
func RandID32() int32 {
	// 使用当前时间作为随机数种子
	rand.Seed(time.Now().UnixNano())
	// 随机生成一个0-4095的整数
	generator := NewID32Generator(int32(rand.Intn(65536)))
	return generator.NewID32()
}

// NewID32Generator 初始化服务器ID
func NewID32Generator(sid int32) *ID32Generator {
	if sid > 65536 {
		panic(fmt.Sprintf("sid %d is out of range", sid))
	}
	serverID32List = append(serverID32List, sid)
	return &ID32Generator{
		serverID:           sid, // 使用 int32
		idCounter:          0,
		mutex:              &sync.Mutex{},
		lastCounterResetTs: 0,
	}
}

// NewID32 获取一个唯一的 int32 ID
func (idGen *ID32Generator) NewID32() int32 {
	now := time.Now().Unix() % 65536 // 使用16位存储时间戳，确保它是 int64
	idGen.mutex.Lock()
	defer idGen.mutex.Unlock()
	if now != idGen.lastCounterResetTs {
		idGen.idCounter = 0
		idGen.lastCounterResetTs = now
	} else {
		idGen.idCounter++
		if idGen.idCounter > 1023 { // 最大1023个ID
			idGen.idCounter = 1023
			log.Panicf("idgen: id counter overflow, serverID: %d, idCounter: %d", idGen.serverID, idGen.idCounter)
		}
	}
	// 将所有位操作的变量转换为 int64，再进行位操作
	return int32((now << 16) | (int64(idGen.serverID) << 4) | int64(idGen.idCounter))
}

// ParseID32 解析一个 int32 id
func ParseID32(id int32) (serverID int32, createAt int64) {
	createAt = int64(id>>16) * 1000
	serverID = (id >> 4) & 4095 // 12位的 serverID
	return serverID, createAt
}

func IsServerID32Valid(id int32) bool {
	return slices.Contains(serverID32List, id)
}
