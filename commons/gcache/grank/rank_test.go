package grank

import (
	"context"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gmongo"
	"go.mongodb.org/mongo-driver/bson"
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

	db, err := gmongo.NewMongoDB(context.Background(), &gmongo.Config{})
	if err != nil {
		return
	}
	err = gmongo.InsertMany(context.Background(), db.GetDB().Collection("test"), docs)
	if err != nil {
		t.Error(err)
	}
	t.Log("insert success")
}

func TestLoadDB(t *testing.T) {
	db, err := gmongo.NewMongoDB(context.Background(), &gmongo.Config{})
	if err != nil {
		return
	}
	lists, err := gmongo.Find[*List](context.Background(), db.GetDB().Collection("test"), bson.D{})
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	t.Logf("load successL:%v", len(lists))
}
