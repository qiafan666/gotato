package grocketmq

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/qiafan666/gotato/commons/gcast"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gface"
	"github.com/qiafan666/gotato/commons/gson"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type TestOrder struct {
}

func NewTestOrder() *TestOrder {
	return &TestOrder{}
}

func (t *TestOrder) Handle(tag, msg string) error {
	fmt.Println(fmt.Sprintf("TestOrder Handle tag:%s msg:%s", tag, msg))
	return nil
}

func TestConsumer(t *testing.T) {

	options := []consumer.Option{
		consumer.WithGroupName("NING_GROUP"),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"10.254.3.87:9876"})),
	}

	consume, err := NewConsumer(context.Background(), gface.NewLogger("grocketmq.consumer", nil), false, options...)
	if err != nil {
		return
	}
	defer consume.Close()

	testOrder := NewTestOrder()
	msgChannel := &MsgChannel{
		Topic: "ning",
	}
	go consume.Consume(context.Background(), msgChannel, testOrder)
	select {}
}

func TestProducer(t *testing.T) {

	options := []producer.Option{
		producer.WithGroupName("NING_GROUP"),
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"10.254.3.87:9876"})),
	}
	produce, err := NewProducer(context.Background(), gface.NewLogger("grocketmq.producer", nil), options...)
	if err != nil {
		return
	}
	defer produce.Close()

	msgChannel := &MsgChannel{
		Topic: "ning",
	}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Second)
		err = produce.Publish(context.Background(), msgChannel, "msg")

	}
	if err != nil {
		return
	}
}

func TestMains(t *testing.T) {
	sig := make(chan os.Signal)
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("SPOT_FEE_FLOW_CONSUMER_GROUP"),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"10.0.0.222:9876"})),
	)
	err := c.Subscribe("SPOT_FEE_FLOW_TOPIC", consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			fmt.Printf("subscribe callback: %v \n", msgs[i])
		}

		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	<-sig
	err = c.Shutdown()
	if err != nil {
		fmt.Printf("shutdown Consumer error: %s", err.Error())
	}
}

func TestAdmin(t *testing.T) {

	// 创建 Admin 客户端
	adm, err := admin.NewAdmin(
		admin.WithResolver(primitive.NewPassthroughResolver([]string{"10.254.3.87:9876"})),
		admin.WithCredentials(primitive.Credentials{
			AccessKey: "admin",
			SecretKey: "admin",
		}))
	if err != nil {
		fmt.Println("Create Admin error:", err)
		return
	}
	defer adm.Close()

	// 获取所有 Topic
	topics, err := adm.FetchAllTopicList(context.Background())
	if err != nil {
		fmt.Println("Fetch Topics error:", err)
		return
	}
	t.Logf("All Topics: %v", topics)

	// 遍历所有 Topic，修改以 "EXCHANGE" 开头的 Topic
	for _, topic := range topics.TopicList {
		if strings.HasPrefix(topic, "EXCHANGE_") {
			fmt.Println("Updating Topic:", topic)

			err = adm.DeleteTopic(context.Background(),
				admin.WithTopicDelete(topic),
			)
			if err != nil {
				fmt.Printf("Delete Topic %s error: %v\n", topic, err)
				continue
			}

			err = adm.CreateTopic(context.Background(), admin.WithTopicCreate(topic),
				admin.WithWriteQueueNums(4),
				admin.WithReadQueueNums(4),
				admin.WithBrokerAddrCreate("172.20.37.209:10911"),
			)
			if err != nil {
				fmt.Printf("Create Topic %s error: %v\n", topic, err)
			} else {
				fmt.Printf("Successfully recreated Topic: %s\n", topic)
			}
		}
	}
}

type updateTopicReq struct {
	ClusterNameList []string `json:"clusterNameList"`
	BrokerNameList  []string `json:"brokerNameList"`
	TopicName       string   `json:"topicName"`
	WriteQueueNums  string   `json:"writeQueueNums"`
	ReadQueueNums   string   `json:"readQueueNums"`
	Perm            int      `json:"perm"`
	Order           bool     `json:"order"`
}

type baseResp struct {
	Status int    `json:"status"`
	Data   bool   `json:"data"`
	ErrMsg string `json:"errMsg"`
}

// dashboard 管理f12接口调用
func TestManager(t *testing.T) {

	// 创建 Admin 客户端
	adm, err := admin.NewAdmin(
		admin.WithResolver(primitive.NewPassthroughResolver([]string{"10.254.3.87:9876"})),
		admin.WithCredentials(primitive.Credentials{
			AccessKey: "admin",
			SecretKey: "admin",
		}))
	if err != nil {
		fmt.Println("Create Admin error:", err)
		return
	}
	defer adm.Close()

	// 获取所有 Topic
	topics, err := adm.FetchAllTopicList(context.Background())
	if err != nil {
		fmt.Println("Fetch Topics error:", err)
		return
	}
	t.Logf("All Topics: %v", topics)

	// 逻辑处理，要修改那些topic
	f := func(allTopic []string) []string {

		return []string{}
	}

	for _, topic := range f(topics.TopicList) {
		req := updateTopicReq{
			ClusterNameList: nil,
			BrokerNameList:  []string{"broker-g0"},
			TopicName:       topic,
			WriteQueueNums:  "1",
			ReadQueueNums:   "1",
			Perm:            6,
			Order:           false,
		}
		fmt.Println(gcast.ToString(req))
		headers := http.Header{
			"Content-Type": []string{"application/json"},
		}
		bodyByte, _, err := gcommon.ProxyRequest(http.MethodPost, headers, "http://10.254.243.245:8080/topic/createOrUpdate.do", gcast.ToByte(req))
		if err != nil {
			t.Errorf("topic %s,request error:%v", req.TopicName, err)
		}

		var resp baseResp
		err = gson.Unmarshal(bodyByte, &resp)
		if err != nil {
			t.Errorf("topic %s,unmarshal error:%v,body:%s", req.TopicName, err, string(bodyByte))
		}

		if resp.Status != 0 || resp.Data != true {
			t.Errorf("topic %s,resp error:%v,body:%s", req.TopicName, resp.ErrMsg, string(bodyByte))
		}
	}
}
