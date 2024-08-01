package gpromise_test

import (
	"github.com/qiafan666/gotato/commons/gpromise"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestCommonFutureAfter(t *testing.T) {
	pm := gpromise.NewManager(1, func() int { return 100 })

	p := pm.NewPromise("testPromise", nil)
	cf1 := gpromise.NewCommonFuture("future1")
	cf2 := gpromise.NewCommonFuture("future2")
	cf3 := gpromise.NewCommonFuture("future3")

	cf1.OnDo = func() error {
		log.Println("cf1 OnDo executed")
		return nil
	}

	cf1.OnCallBack = func(args []interface{}) error {
		log.Println("cf1 OnCallBack executed:" + args[0].(string))
		return nil
	}

	cf2.OnDo = func() error {
		log.Println("cf2 OnDo executed")
		return nil
	}

	cf2.OnCallBack = func(args []interface{}) error {
		log.Println("cf2 OnCallBack executed:" + args[0].(string))
		return nil
	}

	cf3.OnDo = func() error {
		log.Println("cf3 OoDo executed")
		return nil
	}

	cf3.OnCallBack = func(args []interface{}) error {
		log.Println("cf3 OnCallBack executed:" + args[0].(string))
		return nil
	}

	cf1.After(cf2).After(cf3)
	p.Push(cf1)

	p.Start()

	pm.Process(gpromise.GetPfId(p.Id, cf1.Id()), []interface{}{"test"}, nil)
	pm.Process(gpromise.GetPfId(p.Id, cf2.Id()), []interface{}{"test"}, nil)
	pm.Process(gpromise.GetPfId(p.Id, cf3.Id()), []interface{}{"test"}, nil)

	assert.Equal(t, gpromise.FutureStatusFinish, cf1.GetStatus())
	assert.Equal(t, gpromise.FutureStatusFinish, cf2.GetStatus())
	assert.Equal(t, gpromise.FutureStatusFinish, cf3.GetStatus())
}
