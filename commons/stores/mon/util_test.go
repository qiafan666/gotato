package mon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx/logtest"
)

func TestFormatAddrs(t *testing.T) {
	tests := []struct {
		addrs  []string
		expect string
	}{
		{
			addrs:  []string{"a", "b"},
			expect: "a,b",
		},
		{
			addrs:  []string{"a", "b", "c"},
			expect: "a,b,c",
		},
		{
			addrs:  []string{},
			expect: "",
		},
		{
			addrs:  nil,
			expect: "",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expect, FormatAddr(test.addrs))
	}
}

func Test_logDuration(t *testing.T) {
	buf := logtest.NewCollector(t)

	buf.Reset()
	logDuration(context.Background(), "foo", "bar", time.Millisecond, nil)
	assert.Contains(t, buf.String(), "foo")
	assert.Contains(t, buf.String(), "bar")

	buf.Reset()
	logDuration(context.Background(), "foo", "bar", time.Millisecond, errors.New("bar"))
	assert.Contains(t, buf.String(), "foo")
	assert.Contains(t, buf.String(), "bar")
	assert.Contains(t, buf.String(), "fail")
}
