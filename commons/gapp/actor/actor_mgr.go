package actor

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gapp/chanrpc"
	"github.com/qiafan666/gotato/commons/gapp/logger"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/stores/redis"
	"runtime"
	"sync"
	"unsafe"
)

type Mgr struct {
	creator    Creator
	Actors     sync.Map // ActorID -> *Actor，actorMgr写，其他goroutine读
	allActorWg sync.WaitGroup
	/*
		集群唯一actor的redis key prefix， 如果为空说明不需要全局唯一
		集群唯一的actor，比如：玩家、ugc
		不需要集群唯一的actor，比如:scene， world1、world2上可以相同actorID（比如地图id）的actor，ktv、dance用的是房间id（本身已经是全局唯一了）作为地图id
	*/
	globalRedisKey string
	redisClient    *redis.Redis
	index          int32 // 当前的actorMgr在那条线上
}

// NewMgr 启动一个actorMgr，通常结合skeleton module使用
func NewMgr(creator Creator, globalRedisKey string, redisClient *redis.Redis, index int32) *Mgr {
	return &Mgr{
		creator:        creator,
		globalRedisKey: globalRedisKey,
		redisClient:    redisClient,
		index:          index,
	}
}

// StartActor 注册并运行一个Actor
// actorID: Actor唯一ID
// initData: 初始化数据，会传给Actor OnInit
// syncInit: 是否同步等待Actor OnInit流程完成
// 如果是集群唯一的actor，则需要抢到分布式锁，且actor不存在
func (m *Mgr) StartActor(actorID int64, initData any, syncInit bool) error {
	if actorID <= 0 {
		return fmt.Errorf("actorId=%d illegal", actorID)
	}
	var redisLock *redis.RedisLock
	// 集群唯一检查
	if m.globalRedisKey != "" {
		if m.redisClient == nil {
			return fmt.Errorf("actorId=%d redis client is nil", actorID)
		}

		redisLock = redis.NewRedisLock(m.redisClient, GenActorLockKeys(m.globalRedisKey, actorID))
		redisLock.SetExpire(1)
		gotLock, gotLockErr := redisLock.Acquire()
		if gotLockErr != nil {
			return fmt.Errorf("actor%d got lock err %v", actorID, gotLockErr)
		}
		if !gotLock {
			return fmt.Errorf("actorId=%d got lock fail %v", actorID, gotLock)
		}
		exist, _ := m.redisClient.Exists(GenActorKeys(m.globalRedisKey, actorID))
		logger.DefaultLogger.DebugF("StartActor actor:%v exist:%v", actorID, exist)
		if exist {
			return fmt.Errorf("actorId=%d already existed in cluster", actorID)
		}
	}
	a := &Actor{
		id:       actorID,
		closeSig: make(chan bool, 1),
		state:    StateNone,
		delegate: m.creator(actorID),
	}
	_, exist := m.Actors.LoadOrStore(actorID, a)
	if exist {
		return fmt.Errorf("actorId=%d already existed", actorID)
	}
	syncInitCh := make(chan error, 1)
	a.wg.Add(1)
	m.allActorWg.Add(1)

	if redisLock != nil {
		_, _ = redisLock.Release()
		_ = m.redisClient.Set(GenActorKeys(m.globalRedisKey, actorID), gcast.ToString(m.index))
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 2048)
				l := runtime.Stack(buf, false)
				b := buf[:l]
				stack := *(*string)(unsafe.Pointer(&b))
				logger.DefaultLogger.ErrorF("actorMgr StartActor panic actor:%v panic:%v stack:%v", actorID, r, stack)
			}
		}()
		defer m.allActorWg.Done()
		defer a.wg.Done()
		defer m.delActor(actorID)
		a.InitAndRun(initData, syncInitCh)
	}()
	if syncInit {
		err := <-syncInitCh
		return err
	}
	return nil
}

// StopActor 终止Actor
// syncWait==true时，表示同步等待终止
func (m *Mgr) StopActor(actorID int64, syncWait bool) {
	defer func() {
		if r := recover(); r != nil {
			m.Actors.Delete(actorID)
		}
	}()
	a := m.GetActor(actorID)
	if a == nil {
		logger.DefaultLogger.DebugF("stop actor[%d] fail", actorID)
		return
	}
	// 删除actor的redis数据
	if m.globalRedisKey != "" && m.redisClient != nil {
		_, err := m.redisClient.Del(GenActorKeys(m.globalRedisKey, actorID))
		logger.DefaultLogger.DebugF("del actor[%d] key result[%v]", actorID, err)
	}
	a.Stop(syncWait)
}

// GetActor 获取指定Actor
func (m *Mgr) GetActor(actorID int64) *Actor {
	v, ok := m.Actors.Load(actorID)
	if !ok {
		return nil
	}
	return v.(*Actor)
}

// GetActorChanSrv 获取Actor对外暴露的 chanrpc.IServer
func (m *Mgr) GetActorChanSrv(actorID int64) chanrpc.IServer {
	a := m.GetActor(actorID)
	if a == nil {
		return nil
	}
	return a.delegate.ChanSrv()
}

// delActor 删除指定Actor
func (m *Mgr) delActor(actorID int64) {
	m.Actors.Delete(actorID)
	logger.DefaultLogger.DebugF("delActor[%d]", actorID)
}

// StopAllActor 终止所有的Actor
// syncWait==true表示同步等待终止完成
func (m *Mgr) StopAllActor(syncWait bool) {
	actorIDs := make([]int64, 0)
	m.Actors.Range(func(_, v any) bool {
		// 为了让Actor可以并发终止，这里syncWait填false，不依次同步等待终止完成，而是最后等待allActorWg完成
		actorIDs = append(actorIDs, v.(*Actor).id)
		v.(*Actor).Stop(false)
		return true
	})
	// 删除redis actor数据
	logger.DefaultLogger.InfoF("StopAllActor actorIDs%v global redis key[%v] redisC[%v]", actorIDs, m.globalRedisKey, m.redisClient != nil)
	if len(actorIDs) > 0 && m.globalRedisKey != "" && m.redisClient != nil {
		keys := make([]string, 0, len(actorIDs))
		for _, actorID := range actorIDs {
			keys = append(keys, GenActorKeys(m.globalRedisKey, actorID))
		}
		_, err := m.redisClient.Del(keys...)
		if err != nil {
			logger.DefaultLogger.ErrorF("StopActor and del redis role actor err: %v", err)
		}
	}

	if syncWait {
		m.allActorWg.Wait()
	}
}

func (m *Mgr) SetCreator(creator Creator) {
	m.creator = creator
}

func GenActorKeys(actorPrefix string, actorID int64) string {
	return gcommon.StrJoin(":", actorPrefix, gcast.ToString(actorID))
}

func GenActorLockKeys(globalRedisKey string, actorID int64) string {
	return gcommon.StrJoin(":", globalRedisKey, gcast.ToString(actorID))
}