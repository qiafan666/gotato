package grank

import (
	"sort"
	"time"
)

// Item 排行榜上的每个条目
type Item struct {
	ID        int64
	Score     int64
	Place     int32 // 排名
	ExtraData any
}

// ItemList 用于排序
type ItemList []*Item

func (r ItemList) Len() int {
	return len(r)
}

func (r ItemList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ItemList) Less(i, j int) bool {
	return r[i].Score > r[j].Score
}

type List struct {
	Typ      string          `bson:"_id"`
	Items    ItemList        // 按顺序保存排行榜所有条目
	ID2Idx   map[int64]int32 // 描述id到 place index映射
	LastTick int64
}

// newRankList 新建排行榜
func newRankList(typ string) *List {
	return &List{
		Typ:    typ,
		Items:  make([]*Item, 0, DefRankCount),
		ID2Idx: make(map[int64]int32, DefRankCount),
	}
}

// Put 插入新的积分
func (l *List) put(id, score int64, extraData any) int32 {
	item, _ := l.getOrNewItem(id, extraData)
	//ranked := l.sortByGrad(score)
	//if ranked {
	//	return
	//}
	if score > item.Score { // 排名上升
		return l.moveUp(id, item.Place-1, score)
	} else if score < item.Score { // 排名下降
		return l.moveDown(id, item.Place-1, score)
	}
	return item.Place
}

// getOrNewItem 获取指定条目，如果不存在，则创建，并附加到最后一名
func (l *List) getOrNewItem(id int64, extraData any) (*Item, bool) {
	// 得到旧信息
	var item *Item
	idx, ok := l.ID2Idx[id]
	if !ok {
		// 创建新的item
		idx = int32(len(l.Items))
		item = &Item{
			ID:        id,
			Score:     minScore,
			Place:     idx + 1, // Place从1开始
			ExtraData: extraData,
		}
		l.Items = append(l.Items, item)
		l.ID2Idx[id] = idx
	} else {
		item = l.Items[idx]
	}
	return item, ok
}

// update 积分变更，传入的是变更值，而非当前值
func (l *List) update(id, delta int64, extraData any) int32 {
	item, ok := l.getOrNewItem(id, extraData)
	var newScore int64
	if ok {
		newScore = item.Score + delta
	} else {
		newScore = delta
	}
	//ranked := l.sortByGrad(newScore)
	//if ranked {
	//	return
	//}
	// 更新Place
	if delta >= 0 { // 排名上升
		return l.moveUp(id, item.Place-1, newScore)
	} // 排名下降
	return l.moveDown(id, item.Place-1, newScore)
}

// moveUp 将元素往前移动，移动到分数相同的最后一个
func (l *List) moveUp(id int64, idx int32, newScore int64) int32 {
	item := l.Items[idx]
	// 将分数低的向后移动
	for ; idx > 0; idx-- {
		prev := l.Items[idx-1] // 取前面一个
		if newScore <= prev.Score {
			break
		}
		l.ID2Idx[prev.ID] = idx
		l.Items[idx] = prev
		prev.Place++
	}
	// 替换
	l.ID2Idx[id] = idx
	l.Items[idx] = item
	item.Score = newScore
	item.Place = idx + 1
	return item.Place
}

// moveDown 将元素往后移动，移动到分数相同的最前一个
func (l *List) moveDown(id int64, idx int32, newScore int64) int32 {
	item := l.Items[idx]
	// 移动比自身分数高的
	for ; idx < int32(len(l.Items))-1; idx++ {
		next := l.Items[idx+1] // 取后面面一个
		if newScore >= next.Score {
			break
		}
		// 向前移动
		l.ID2Idx[next.ID] = idx
		l.Items[idx] = next
		next.Place--
	}
	l.ID2Idx[id] = idx
	l.Items[idx] = item
	item.Score = newScore
	item.Place = idx + 1
	return item.Place
}

// Get 获取指定玩家的排名
func (l *List) get(id int64) *Item {
	idx, ok := l.ID2Idx[id]
	if !ok {
		return nil
	}
	return l.Items[idx]
}

// del 删除指定玩家的信息
func (l *List) del(id int64) {
	idx, ok := l.ID2Idx[id]
	if !ok {
		return
	}
	delete(l.ID2Idx, id)
	l.Items = append(l.Items[:idx], l.Items[idx+1:]...)
	// 更新后续玩家的place位置
	for _, p := range l.Items[idx:] {
		p.Place--
		l.ID2Idx[p.ID] = p.Place - 1
	}
}

// Sort 主动排序
func (l *List) sort() {
	// 排序
	sort.Stable(l.Items)
	// 更新index,place
	for i, item := range l.Items {
		item.Place = int32(i + 1)
		l.ID2Idx[item.ID] = int32(i)
	}
}

// SeePage 查看某页，相同积分同一个名次
func (l *List) seePage(index, length int32) []*Item {
	if index < 0 {
		index = 0
	}
	if index >= int32(len(l.Items)) {
		return nil
	}
	if index+length >= int32(len(l.Items)) {
		length = int32(len(l.Items)) - index
	}
	return l.Items[index : index+length]
}

// isSort 判断是否排序正确
func (l *List) isSort() bool {
	for i, item := range l.Items[:immediateRankCount] {
		if l.ID2Idx[item.ID] != int32(i) {
			return false
		}
	}
	return true
}

// getByPlace 获取指定排名的条目
func (l *List) getByPlace(place int32) *Item {
	if place > int32(len(l.Items)) {
		return nil
	}
	return l.Items[place-1]
}

// sortByGrad 定时主动排序一次
func (l *List) sortByGrad(score int64) bool {
	var flag int64
	if len(l.Items) < immediateRankCount {
		return false
	}
	item := l.Items[immediateRankCount-1]
	if item != nil {
		flag = item.Score
	}
	if score < flag {
		now := time.Now().Unix()
		if now-rankGrad < l.LastTick {
			return true
		}
		l.LastTick = now
		l.sort()
		return true
	}
	return false
}
