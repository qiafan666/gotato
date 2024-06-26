package utils

import "testing"

func TestBusSync(t *testing.T) {
	//同步测试
	bus := NewMsgBus("test", true, false, false)

	err := bus.Subscribe(func(a, b int) {
		t.Logf("a+b=%d", a+b)
	})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, i)
		if err != nil {
			t.Error()
		}
	}
}

func TestBusAsync(t *testing.T) {
	//异步串行测试
	bus := NewMsgBus("test", false, true, false)
	err := bus.Subscribe(func(a, b int) {
		t.Logf("a+b=%d", a+b)
	})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, i)
		if err != nil {
			t.Error()
		}
	}

	bus.WaitAsync()
}

func TestBusAsync2(t *testing.T) {
	//异步并行测试
	bus := NewMsgBus("test", false, false, false)
	err := bus.Subscribe(func(a, b int) {
		t.Logf("a+b=%d", a+b)
	})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, i)
		if err != nil {
			t.Error()
		}
	}

	//等待全部执行，否则线程可能没执行完
	bus.WaitAsync()
}

func TestBuss(t *testing.T) {
	//测试多个订阅者
	bus := NewMsgBus("test", true, false, false)
	err := bus.Subscribe(func(a, b int) {
		t.Logf("sub1,a+b=%d", a+b)
	})
	err = bus.Subscribe(func(a, b int) {
		t.Logf("sub2,a+b=%d", a+b)
	})
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, i)
		if err != nil {
			t.Error()
		}
	}
}

func TestOnce(t *testing.T) {
	//测试一次性订阅
	bus := NewMsgBus("test", true, false, true)
	err := bus.Subscribe(func(a, b int) {
		t.Logf("a+b=%d", a+b)
	})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, i)
		if err != nil {
			t.Error()
		}
	}
}

func TestError(t *testing.T) {
	//测试错误提示
	bus := NewMsgBus("test", true, false, false)
	err := bus.Subscribe(
		func(a, b int) {
			t.Logf("a+b=%d", a+b)
		})
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		err = bus.Publish(1, "error")
		if err != nil {
			t.Error(err)
		}
	}
}
