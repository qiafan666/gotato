package gpool

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	p, err := NewPool[*testItem](context.Background(), &Options[*testItem]{
		MaxSize:  10,
		InitSize: 3,
		New:      newTestItem,
	})
	assert.Nil(t, err, err)

	wg := new(sync.WaitGroup)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			item, err := p.Get(context.Background())
			assert.Nil(t, err, err)

			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(10)
			if r%10 == 0 {
				//p.closeItem(item)
				//return
			}

			p.Put(item)
			size := len(p.items)
			fmt.Printf("size: %d\n", size)
			assert.LessOrEqual(t, size, 10)
		}()
	}

	wg.Wait()
}

var (
	id int
	mu sync.Mutex
)

type testItem struct {
	id              int
	closed          bool
	closeChanNotify chan any
	closeOnce       sync.Once
}

func newTestItem() (*testItem, error) {
	mu.Lock()
	defer mu.Unlock()
	id++
	fmt.Printf("new item, id:%d\n", id)
	return &testItem{
		id:              id,
		closeChanNotify: make(chan any),
	}, nil
}

func (i *testItem) Close() error {
	i.closeOnce.Do(func() {
		fmt.Printf("close item, id:%d\n", i.id)
		i.closed = true
		i.closeChanNotify <- i.id
	})
	return nil
}

func (i *testItem) IsClosed() bool {
	return i.closed
}

func (i *testItem) CloseNotify() <-chan any {
	fmt.Printf("close notify item, id:%d\n", i.id)
	return i.closeChanNotify
}
