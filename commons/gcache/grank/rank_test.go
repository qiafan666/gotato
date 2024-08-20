package grank

import (
	"github.com/qiafan666/gotato/commons/gcast"
	"math/rand"
	"testing"
)

// 测试排序正确性
func TestMgr_Put(t *testing.T) {
	// 初始化
	mgr := NewRankListMgr()
	mgr.NewRankList("1")
	// 测试
	for i := 0; i < 100000; i++ {
		score := int64(rand.Intn(10000))
		id := rand.Int63n(200)
		mgr.Put("1", id, score, nil)
	}
	if !mgr.GetRankList("1").isSort() {
		t.Error("sort rank list err")
	}
	t.Log("check ok")
}

// 排序性能
func BenchmarkRankList(b *testing.B) {
	b.StopTimer()
	// 初始化
	mgr := NewRankListMgr()
	mgr.NewRankList("1")
	roleCount := 10000 // 玩家数量
	interval := 100    // 分数变更间隔
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		id := rand.Int63n(int64(roleCount))
		score := int64(rand.Intn(2*interval) - interval)
		item := mgr.Get("1", id)
		if item != nil {
			score += item.Score
		}
		mgr.Put("1", id, score, nil)
	}
}

func TestInsertDB(t *testing.T) {
	mgr := NewRankListMgr()
	// 测试
	for i := 0; i < 100000; i++ {
		score := int64(rand.Intn(10000))
		id := rand.Int63n(200)
		mgr.Put(gcast.ToString(rand.Intn(10)), id, score, nil)
	}

	docs := make([]interface{}, 0, 10)
	for _, rank := range mgr.Ranks {
		docs = append(docs, rank)
	}
	//_, err := gmongo.InsertMany("mongodb://127.0.0.1:27017/", "test", "rank_test", docs)
	//if err != nil {
	//	t.Error("err")
	//}
	t.Log("insert success")
}

func TestLoadDB(t *testing.T) {
	//cursor, err := gmongo.Find("mongodb://127.0.0.1:27017/", "meta", "rank_test", bson.D{})
	//if err != nil {
	//	t.Errorf("%v", err)
	//	return
	//}
	//ranks := make([]*List, 0)
	//if err = cursor.All(context.Background(), &ranks); err != nil {
	//	t.Errorf("%v", err)
	//}
	//t.Logf("load successL:%v", len(ranks))
}
