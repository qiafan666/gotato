package gid

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	"math/rand"
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

// serverID 服务器ID
var serverID int64

// CounterBits ID编码中自增器所占位数
const CounterBits = 20

// ServerIdBits ServerId所占位数
const ServerIdBits = 12

// MaxCounterInSecond 单秒允许生成的最大ID数量
const MaxCounterInSecond = (1 << CounterBits) - 1

// 2024-01-01 00:00:00 标准时间，不得修改
const epoch = 1704038400

// NewServerID 初始化服务器ID
func NewServerID(sid int) {
	maxID := 1<<(ServerIdBits) - 1
	if sid > maxID {
		panic(fmt.Sprintf("Init fail; sid %d must be lower or equal than %d", sid, maxID))
	}
	serverID = int64(sid)
}

// 以下全局变量仅被NewID()使用，并且被mutex保证互斥访问
// 互斥访问 idCounter 和 lastCounterResetTs
var mutex sync.Mutex

// idCounter ID计数器
var idCounter = int64(0)

// 上一次Counter重置时的时间戳
var lastCounterResetTs int64

// ID 获取一个唯一ID
func ID() int64 {
	now := time.Now().Unix()
	mutex.Lock()
	defer mutex.Unlock()
	if now != lastCounterResetTs {
		idCounter = 0
		lastCounterResetTs = now
	} else {
		idCounter++
		// 单秒生成的ID超过最大限制，视为严重Bug
		// 1. 线上环境，DPanic不会让程序直接panic，避免不严重的overflow引起严重的问题
		// 2. 发生overflow，idCounter=MaxCounterInSecond 继续运行，便于后续找到出问题的ID
		if idCounter > MaxCounterInSecond {
			idCounter = MaxCounterInSecond
		}
	}
	return (now-epoch)<<(ServerIdBits+CounterBits) | serverID<<CounterBits | idCounter
}

// ParseID 解析一个id
func ParseID(id int64) (serverID int32, createAt int64) {
	createAt = (id >> (ServerIdBits + CounterBits)) + epoch
	serverID = int32(uint32(id) >> CounterBits)
	return serverID, createAt
}

// RandID( 生成一个随机ID serverID 随机
func RandID() int64 {
	now := time.Now().Unix()
	mutex.Lock()
	defer mutex.Unlock()
	if now != lastCounterResetTs {
		idCounter = 0
		lastCounterResetTs = now
	} else {
		idCounter++
		// 单秒生成的ID超过最大限制，视为严重Bug
		// 1. 线上环境，DPanic不会让程序直接panic，避免不严重的overflow引起严重的问题
		// 2. 发生overflow，idCounter=MaxCounterInSecond 继续运行，便于后续找到出问题的ID
		if idCounter > MaxCounterInSecond {
			idCounter = MaxCounterInSecond
		}
	}
	randomServerID := gcast.ToInt64(rand.Intn(1<<(ServerIdBits) - 1))
	return (now-epoch)<<(ServerIdBits+CounterBits) | randomServerID<<CounterBits | idCounter
}
