package gbatcher

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gcommon"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestBatcher(t *testing.T) {

	config := Config{
		size:       500,
		worker:     50,
		interval:   100 * time.Millisecond,
		dataBuffer: 500,
		syncWait:   true,
		buffer:     50,
	}
	b := New[string](
		WithSize(config.size),
		WithWorker(config.worker),
		WithInterval(config.interval),
		WithDataBuffer(config.dataBuffer),
		WithSyncWait(config.syncWait),
		WithBuffer(config.buffer),
	)

	b.Do = func(ctx context.Context, channelID int, vals *Msg[string]) {
		t.Logf("Channel %d Processed batch: %v", channelID, vals)
	}
	b.OnComplete = func(lastMessage *string, totalCount int) {
		t.Logf("Completed processing with last message: %v, total count: %d", *lastMessage, totalCount)
	}

	//分片函数，将数据分片到不同的worker中进行处理
	b.Sharding = func(key string) int {
		hashCode := gcommon.Str2Uint32(key)
		return (int(hashCode) + rand.Intn(config.worker)) % config.worker
	}

	//分组函数，可以根据数据中的某些特性进行分组，比如用户id，商品id等，方便进行统一事务处理
	b.Key = func(data *string) string {
		return gcast.ToString(gcast.ToInt(strings.Split(*data, "-")[1]) % 10)
	}

	err := b.Start()
	if err != nil {
		t.Fatal(err)
	}

	// Test normal data processing
	for i := 0; i < 10000; i++ {
		data := "data" + fmt.Sprintf("-%d", i)
		if err := b.Put(context.Background(), &data); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Duration(1) * time.Second)
	start := time.Now()
	// Wait for all processing to finish
	b.Close()

	elapsed := time.Since(start)
	t.Logf("Close took %s", elapsed)

	if len(b.data) != 0 {
		t.Error("Data channel should be empty after closing")
	}
}
