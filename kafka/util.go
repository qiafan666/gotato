package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
)

// 定义一个唯一的上下文键类型，避免键冲突
type mqHeaderKeyType struct{}

// SetContextWithMQHeader 将任意的消息队列 headers 设置到上下文中
func SetContextWithMQHeader(ctx context.Context, headers []sarama.RecordHeader) context.Context {
	// 创建一个 map 来存储 headers 的键值对
	headerMap := make(map[string]string)

	// 遍历传入的 headers，将键值对存入 map 中
	for _, header := range headers {
		headerMap[string(header.Key)] = string(header.Value)
	}

	// 使用 context.WithValue 将 headerMap 存储到上下文中
	return context.WithValue(ctx, mqHeaderKeyType{}, headerMap)
}

// GetMQHeaderFromContext 从上下文中检索消息队列 headers
func GetMQHeaderFromContext(ctx context.Context) (map[string]string, bool) {
	// 从上下文中取出 headers 的 map
	headers, ok := ctx.Value(mqHeaderKeyType{}).(map[string]string)
	return headers, ok
}

// GetMQHeaderWithContext 从上下文中提取 Kafka 消息队列的 headers
func GetMQHeaderWithContext(ctx context.Context) ([]sarama.RecordHeader, error) {
	// 从上下文中获取已存储的 headers
	headersMap, ok := GetMQHeaderFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no MQ headers found in context")
	}

	// 将 map 转换为 []sarama.RecordHeader 格式
	var headers []sarama.RecordHeader
	for key, value := range headersMap {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	return headers, nil
}
