package grank

import (
	"math"
)

const (
	minScore           = math.MinInt64
	DefRankCount       = 1000 // 默认的排行榜条目
	immediateRankCount = 100  // 前面的实时更新，后面的懒惰更新
	rankGrad           = 60   // 60秒定时更新一次
)

// Mgr 排行榜管理器
type Mgr struct {
	Ranks map[string]*List
}

// NewRankListMgr 创建排行榜管理器
func NewRankListMgr() *Mgr {
	return &Mgr{
		Ranks: make(map[string]*List),
	}
}

// NewRankList 新建排行榜
func (mgr *Mgr) NewRankList(typ string) {
	if _, ok := mgr.Ranks[typ]; ok {
		return
	}
	mgr.Ranks[typ] = newRankList(typ)
}

// ClearRankList 清除指定排行榜
func (mgr *Mgr) ClearRankList(typ string) {
	delete(mgr.Ranks, typ)
}

// GetRankList 获取指定排行榜
func (mgr *Mgr) GetRankList(typ string) *List {
	return mgr.Ranks[typ]
}

// Put 插入指定玩家排名数据，传入的是当前分数
func (mgr *Mgr) Put(typ string, id, score int64, extraData any) int32 {
	rank := mgr.Ranks[typ]
	if rank == nil {
		rank = newRankList(typ)
		mgr.Ranks[typ] = rank
	}
	return rank.put(id, score, extraData)
}

// Update 更新指定玩家排名数据，传入的是分数变更
func (mgr *Mgr) Update(typ string, id, delta int64, extraData any) int32 {
	rank := mgr.Ranks[typ]
	if rank == nil {
		rank = newRankList(typ)
		mgr.Ranks[typ] = rank
	}
	return rank.update(id, delta, extraData)
}

// Total 总条目
func (mgr *Mgr) Total(typ string) int32 {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return 0
	}
	return int32(len(rank.ID2Idx))
}

// Del 删除指定玩家排名数据
func (mgr *Mgr) Del(typ string, id int64) {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return
	}
	rank.del(id)
}

// GetByPlace 获取指定排名的条目
func (mgr *Mgr) GetByPlace(typ string, place int32) *Item {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return nil
	}
	return rank.getByPlace(place)
}

// Get 获取指定玩家排名数据
func (mgr *Mgr) Get(typ string, id int64) *Item {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return nil
	}
	return rank.get(id)
}

// Sort 强制重排排行榜
func (mgr *Mgr) Sort(typ string) {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return
	}
	rank.sort()
}

// GetByPage 查看某个排行榜的指定页
func (mgr *Mgr) GetByPage(typ string, start, stop int32) []*Item {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return nil
	}
	// place从1开始, index从0开始
	return rank.seePage(start-1, stop-start)
}

// ForeachFunForRankList 遍历排行榜的函数
type ForeachFunForRankList func(ele *Item)

// Foreach 遍历排行榜
func (mgr *Mgr) Foreach(typ string, doFun ForeachFunForRankList) bool {
	rank := mgr.Ranks[typ]
	if rank == nil {
		return false
	}
	for _, user := range rank.Items {
		doFun(user)
	}
	return true
}
