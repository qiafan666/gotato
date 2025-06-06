package gbatcher

import (
	"context"
	"fmt"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gerr"
	"strings"
	"sync"
	"time"
)

var (
	DefaultDataChanSize = 1000
	DefaultSize         = 100
	DefaultBuffer       = 100
	DefaultWorker       = 5
	DefaultInterval     = time.Second
)

type Config struct {
	size     int           // 消息聚合的数量（每个批次的最大数据量）
	buffer   int           // 单个协程中运行的缓存数量（每个工作协程通道的缓冲区大小）
	worker   int           // 并行处理的工作协程数量
	interval time.Duration // 消息聚合的时间间隔（批次处理的时间间隔）
	syncWait bool          // 分发消息后是否同步等待消息被消费完成
}

type Option func(c *Config)

func WithSize(s int) Option {
	return func(c *Config) {
		c.size = s
	}
}

func WithBuffer(b int) Option {
	return func(c *Config) {
		c.buffer = b
	}
}

func WithWorker(w int) Option {
	return func(c *Config) {
		c.worker = w
	}
}

func WithInterval(i time.Duration) Option {
	return func(c *Config) {
		c.interval = i
	}
}

func WithSyncWait(wait bool) Option {
	return func(c *Config) {
		c.syncWait = wait
	}
}

type Batcher[T any] struct {
	config *Config

	globalCtx  context.Context
	cancel     context.CancelFunc
	Do         func(ctx context.Context, channelID int, val *Msg[T])
	OnComplete func(lastMessage *T, totalCount int)
	Sharding   func(key string) int
	Key        func(data *T) string
	HookFunc   func(triggerID string, messages map[string][]*T, totalCount int, lastMessage *T)
	data       chan *T
	chArrays   []chan *Msg[T]
	wait       sync.WaitGroup
	counter    sync.WaitGroup
}

func emptyOnComplete[T any](*T, int) {}
func emptyHookFunc[T any](string, map[string][]*T, int, *T) {
}

func New[T any](opts ...Option) *Batcher[T] {
	b := &Batcher[T]{
		OnComplete: emptyOnComplete[T],
		HookFunc:   emptyHookFunc[T],
	}
	config := &Config{
		size:     DefaultSize,
		buffer:   DefaultBuffer,
		worker:   DefaultWorker,
		interval: DefaultInterval,
	}
	for _, opt := range opts {
		opt(config)
	}
	b.config = config
	b.data = make(chan *T, DefaultDataChanSize)
	b.globalCtx, b.cancel = context.WithCancel(context.Background())

	b.chArrays = make([]chan *Msg[T], b.config.worker)
	for i := 0; i < b.config.worker; i++ {
		b.chArrays[i] = make(chan *Msg[T], b.config.buffer)
	}
	return b
}

func (b *Batcher[T]) Worker() int {
	return b.config.worker
}

func (b *Batcher[T]) Start() error {
	if b.Sharding == nil {
		return gerr.New("Sharding function is required").Wrap()
	}
	if b.Do == nil {
		return gerr.New("Do function is required").Wrap()
	}
	if b.Key == nil {
		return gerr.New("Key function is required").Wrap()
	}
	b.wait.Add(b.config.worker)
	for i := 0; i < b.config.worker; i++ {
		go b.run(i, b.chArrays[i])
	}
	b.wait.Add(1)
	go b.scheduler()
	return nil
}

func (b *Batcher[T]) Put(ctx context.Context, data *T) error {
	if data == nil {
		return gerr.New("data can not be nil").Wrap()
	}
	select {
	case <-b.globalCtx.Done():
		return gerr.New("data channel is closed").Wrap()
	case <-ctx.Done():
		return ctx.Err()
	case b.data <- data:
		return nil
	}
}

func (b *Batcher[T]) scheduler() {
	ticker := time.NewTicker(b.config.interval)
	defer func() {
		ticker.Stop()
		for _, ch := range b.chArrays {
			close(ch)
		}
		close(b.data)
		b.wait.Done()
	}()

	vals := make(map[string][]*T)
	count := 0
	var lastAny *T

	for {
		select {
		case data, ok := <-b.data:
			if !ok {
				// If the data channel is closed unexpectedly
				return
			}
			if data == nil {
				if count > 0 {
					b.distributeMessage(vals, count, lastAny)
				}
				return
			}

			key := b.Key(data)
			vals[key] = append(vals[key], data)
			lastAny = data

			count++
			if count >= b.config.size {

				b.distributeMessage(vals, count, lastAny)
				vals = make(map[string][]*T)
				count = 0
			}

		case <-ticker.C:
			if count > 0 {

				b.distributeMessage(vals, count, lastAny)
				vals = make(map[string][]*T)
				count = 0
			}
		}
	}
}

type Msg[T any] struct {
	key       string
	triggerID string
	val       []*T
}

func (m Msg[T]) Key() string {
	return m.key
}

func (m Msg[T]) TriggerID() string {
	return m.triggerID
}

func (m Msg[T]) Val() []*T {
	return m.val
}

func (m Msg[T]) String() string {
	var sb strings.Builder
	sb.WriteString("Key: ")
	sb.WriteString(m.key)
	sb.WriteString(", Values: [")
	for i, v := range m.val {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", *v))
	}
	sb.WriteString("]")
	return sb.String()
}

func (b *Batcher[T]) distributeMessage(messages map[string][]*T, totalCount int, lastMessage *T) {
	triggerID := gcommon.GenerateUUID()
	b.HookFunc(triggerID, messages, totalCount, lastMessage)
	for key, data := range messages {
		if b.config.syncWait {
			b.counter.Add(1)
		}
		channelID := b.Sharding(key)
		b.chArrays[channelID] <- &Msg[T]{key: key, triggerID: triggerID, val: data}
	}
	if b.config.syncWait {
		b.counter.Wait()
	}
	b.OnComplete(lastMessage, totalCount)
}

func (b *Batcher[T]) run(channelID int, ch <-chan *Msg[T]) {
	defer b.wait.Done()
	for {
		select {
		case messages, ok := <-ch:
			if !ok {
				return
			}
			b.Do(context.Background(), channelID, messages)
			if b.config.syncWait {
				b.counter.Done()
			}
		}
	}
}

func (b *Batcher[T]) Close() {
	b.cancel() // Signal to stop put data
	b.data <- nil
	//wait all goroutines exit
	b.wait.Wait()
}
