package redis

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	red "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/logx/logtest"
	ztrace "github.com/zeromicro/go-zero/core/trace"
	tracesdk "go.opentelemetry.io/otel/trace"
)

func TestHookProcessCase1(t *testing.T) {
	ztrace.StartAgent(ztrace.Config{
		Name:     "go-zero-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer ztrace.StopAgent()

	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx, err := durationHook.BeforeProcess(context.Background(), red.NewCmd(context.Background()))
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background())))
	assert.False(t, strings.Contains(buf.String(), "slow"))
	assert.Equal(t, "redis", tracesdk.SpanFromContext(ctx).(interface{ Name() string }).Name())
}

func TestHookProcessCase2(t *testing.T) {
	ztrace.StartAgent(ztrace.Config{
		Name:     "go-zero-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer ztrace.StopAgent()

	w := logtest.NewCollector(t)

	ctx, err := durationHook.BeforeProcess(context.Background(), red.NewCmd(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "redis", tracesdk.SpanFromContext(ctx).(interface{ Name() string }).Name())

	time.Sleep(slowThreshold.Load() + time.Millisecond)

	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background(), "foo", "bar")))
	assert.True(t, strings.Contains(w.String(), "slow"))
	assert.True(t, strings.Contains(w.String(), "trace"))
	assert.True(t, strings.Contains(w.String(), "span"))
}

func TestHookProcessCase3(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	assert.Nil(t, durationHook.AfterProcess(context.Background(), red.NewCmd(context.Background())))
	assert.True(t, buf.Len() == 0)
}

func TestHookProcessCase4(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background())))
	assert.True(t, buf.Len() == 0)
}

func TestHookProcessPipelineCase1(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	_, err := durationHook.BeforeProcessPipeline(context.Background(), []red.Cmder{})
	assert.NoError(t, err)
	ctx, err := durationHook.BeforeProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	})
	assert.NoError(t, err)
	assert.Equal(t, "redis", tracesdk.SpanFromContext(ctx).(interface{ Name() string }).Name())

	assert.NoError(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{}))
	assert.NoError(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.False(t, strings.Contains(buf.String(), "slow"))
}

func TestHookProcessPipelineCase2(t *testing.T) {
	ztrace.StartAgent(ztrace.Config{
		Name:     "go-zero-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer ztrace.StopAgent()

	w := logtest.NewCollector(t)

	ctx, err := durationHook.BeforeProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	})
	assert.NoError(t, err)
	assert.Equal(t, "redis", tracesdk.SpanFromContext(ctx).(interface{ Name() string }).Name())

	time.Sleep(slowThreshold.Load() + time.Millisecond)

	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background(), "foo", "bar"),
	}))
	assert.True(t, strings.Contains(w.String(), "slow"))
	assert.True(t, strings.Contains(w.String(), "trace"))
	assert.True(t, strings.Contains(w.String(), "span"))
}

func TestHookProcessPipelineCase3(t *testing.T) {
	w := logtest.NewCollector(t)

	assert.Nil(t, durationHook.AfterProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.True(t, len(w.String()) == 0)
}

func TestHookProcessPipelineCase4(t *testing.T) {
	w := logtest.NewCollector(t)

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.True(t, len(w.String()) == 0)
}

func TestHookProcessPipelineCase5(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.True(t, buf.Len() == 0)
}

func TestLogDuration(t *testing.T) {
	w := logtest.NewCollector(t)

	logDuration(context.Background(), []red.Cmder{
		red.NewCmd(context.Background(), "get", "foo"),
	}, 1*time.Second)
	assert.True(t, strings.Contains(w.String(), "get foo"))

	logDuration(context.Background(), []red.Cmder{
		red.NewCmd(context.Background(), "get", "foo"),
		red.NewCmd(context.Background(), "set", "bar", 0),
	}, 1*time.Second)
	assert.True(t, strings.Contains(w.String(), `get foo\nset bar 0`))
}

func TestFormatError(t *testing.T) {
	// Test case: err is OpError
	err := &net.OpError{
		Err: mockOpError{},
	}
	assert.Equal(t, "timeout", formatError(err))

	// Test case: err is nil
	assert.Equal(t, "", formatError(nil))

	// Test case: err is red.Nil
	assert.Equal(t, "", formatError(red.Nil))

	// Test case: err is io.EOF
	assert.Equal(t, "eof", formatError(io.EOF))

	// Test case: err is context.DeadlineExceeded
	assert.Equal(t, "context deadline", formatError(context.DeadlineExceeded))

	// Test case: err is breaker.ErrServiceUnavailable
	assert.Equal(t, "breaker", formatError(breaker.ErrServiceUnavailable))

	// Test case: err is unknown
	assert.Equal(t, "unexpected error", formatError(errors.New("some error")))
}

type mockOpError struct {
}

func (mockOpError) Error() string {
	return "mock error"
}

func (mockOpError) Timeout() bool {
	return true
}
