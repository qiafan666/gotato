package gticker

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	NewTicker(1*time.Second, func() {
		t.Log("tick")
	}).Run(context.Background())
}
