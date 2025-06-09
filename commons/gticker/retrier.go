package gticker

import (
	"context"
	"time"
)

// Retry 封装一个带重试机制的执行器
type Retry struct {
	interval time.Duration
	action   func() error
	retryNum int
}

// NewRetry 创建一个带重试的执行器
func NewRetry(interval time.Duration, action func() error) *Retry {
	return &Retry{
		interval: interval,
		action:   action,
	}
}

// SetRetryNum 设置重试次数
func (r *Retry) SetRetryNum(num int) {
	r.retryNum = num
}

// Run 开始运行，直到成功或 ctx 被取消
func (r *Retry) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		err := r.action()
		if err == nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 继续重试
		}
	}
}
