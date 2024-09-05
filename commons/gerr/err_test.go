package gerr

import "testing"

func TestError(t *testing.T) {
	testErr := NewCodeError(10000, "test error", "")
	t.Log(testErr.Error()) // output: test error

	testErr1 := NewCodeError(10001, "test error1", "").Wrap()
	t.Log(testErr1.Error()) // output: test error1

	testErr2 := NewCodeError(10002, "test error2", "").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr2.Error()) // output: wrap msg, key=value: test error2

	err := Unwrap(testErr2)
	if err != nil {
		t.Log(err.Error()) // output: wrap msg, key=value: test error2
	}

	testErr3 := NewCodeError(10003, "test error3", "").WithDetail("msg detail")
	t.Log(testErr3.Error()) // output: test error3 ;detail=msg detail

	err = Unwrap(testErr3)
	if err == nil {
		return
	}
	t.Log(err.Error()) // output: test error3 ;detail=msg detail

	testErr4 := New("test error4", "key", "value")

	t.Log(testErr4.Error()) // output: test error4, key=value

	testErr5 := New("test error5", "key", "value").Wrap()
	t.Log(testErr5.Error()) // output: test error5, key=value

	testErr6 := New("test error6", "key", "value").WrapMsg("wrap msg", "key", "value")
	t.Log(testErr6.Error()) // output: wrap msg, key=value: test error6, key=value
}
