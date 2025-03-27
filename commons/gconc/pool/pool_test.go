package pool_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/commons/gconc/pool"
	"testing"
	"time"
)

func TestExamplePool(t *testing.T) {
	p := pool.New().WithMaxGoroutines(3)
	for i := 0; i < 5; i++ {
		p.Go(func() {
			fmt.Println("conc")
		})
	}
	p.Wait()
	// Output:
	// conc
	// conc
	// conc
	// conc
	// conc
}

func TestCancelPool(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	// 使用 conc.WithContext 让所有 goroutine 受 `ctx` 控制
	group := pool.New().WithContext(ctx)

	// 启动多个任务
	for i := 0; i < 3; i++ {
		i := i
		group.Go(func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					fmt.Println("任务被取消:", ctx.Err())
					return ctx.Err()
				default:
					fmt.Printf("任务运行%d\n", i)
					time.Sleep(500 * time.Millisecond) // 模拟任务

				}
			}
		})
	}

	// 3 秒后取消所有任务
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("任务开始取消...")
		cancel()
	}()

	// 等待所有任务完成
	if err := group.Wait(); err != nil {
		fmt.Println("任务执行中止:", err)
	} else {
		fmt.Println("所有任务完成")
	}
}

func TestExampleErrorPool(t *testing.T) {
	p := pool.New().WithErrors()
	for i := 0; i < 3; i++ {
		i := i
		p.Go(func() error {
			if i == 2 {
				return errors.New("oh no!")
			}
			return nil
		})
	}
	err := p.Wait()
	fmt.Println(err)
	// Output:
	// oh no!
}

func TestExampleResultPool(t *testing.T) {
	p := pool.NewWithResults[int]()
	for i := 0; i < 10; i++ {
		i := i
		p.Go(func() int {
			return i * 2
		})
	}
	res := p.Wait()
	fmt.Println(res)

	// Output:
	// [0 2 4 6 8 10 12 14 16 18]
}
