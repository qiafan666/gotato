package ges

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

// IEs 定义了需要实现的方法的接口
type IEs interface {
	Client() *elastic.Client
	Search() *elastic.SearchService
	All(ctx context.Context, index string, value interface{}) error
	GetById(ctx context.Context, index, id string) (*elastic.GetResult, error)
	DeleteById(ctx context.Context, index, id string) (*elastic.DeleteResponse, error)
	UpdateById(ctx context.Context, index, id string, doc interface{}) (*elastic.UpdateResponse, error)
	QueryByTerm(ctx context.Context, index, field, value string, result interface{}) error
	QueryByMatch(ctx context.Context, index, field, value string, result interface{}) error
	QueryByRange(ctx context.Context, index, field string, from, to interface{}, result interface{}) error
	QueryByBool(ctx context.Context, index string, must, should, mustNot []elastic.Query, result interface{}) error
	QueryByPhrase(ctx context.Context, index, field, phrase string, result interface{}) error
}

// EsClient 实现了 EsClient 接口
type esClient struct {
	client *elastic.Client
}

// NewEs 创建一个新的 Elasticsearch 客户端并返回接口
func NewEs(options ...elastic.ClientOptionFunc) (IEs, error) {
	client, err := elastic.NewClient(options...)
	if err != nil {
		return nil, err
	}
	return &esClient{client: client}, nil
}

func (e *esClient) Client() *elastic.Client {
	return e.client
}

func (e *esClient) Search() *elastic.SearchService {
	return e.client.Search()
}

// All 查询索引中的所有文档并解析到 value 中
func (e *esClient) All(ctx context.Context, index string, value interface{}) error {
	query := elastic.NewMatchAllQuery()
	result, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	if result.Hits.TotalHits.Value > 0 {
		for _, hit := range result.Hits.Hits {
			err = json.Unmarshal(hit.Source, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetById 根据索引和ID获取文档
func (e *esClient) GetById(ctx context.Context, index, id string) (*elastic.GetResult, error) {
	return e.client.Get().Index(index).Id(id).Do(ctx)
}

// DeleteById 根据索引和ID删除文档
func (e *esClient) DeleteById(ctx context.Context, index, id string) (*elastic.DeleteResponse, error) {
	return e.client.Delete().Index(index).Id(id).Do(ctx)
}

// UpdateById 根据索引和ID更新文档
func (e *esClient) UpdateById(ctx context.Context, index, id string, doc interface{}) (*elastic.UpdateResponse, error) {
	return e.client.Update().Index(index).Id(id).Doc(doc).Do(ctx)
}

// QueryByTerm 根据字段值精确查询文档
func (e *esClient) QueryByTerm(ctx context.Context, index, field, value string, result interface{}) error {
	query := elastic.NewTermQuery(field, value)
	searchResult, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	return parseSearchResult(searchResult, result)
}

// QueryByMatch 根据字段值模糊查询文档
func (e *esClient) QueryByMatch(ctx context.Context, index, field, value string, result interface{}) error {
	query := elastic.NewMatchQuery(field, value)
	searchResult, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	return parseSearchResult(searchResult, result)
}

// QueryByRange 根据范围查询文档
func (e *esClient) QueryByRange(ctx context.Context, index, field string, from, to interface{}, result interface{}) error {
	query := elastic.NewRangeQuery(field).From(from).To(to)
	searchResult, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	return parseSearchResult(searchResult, result)
}

// QueryByBool 使用布尔查询条件查询文档
func (e *esClient) QueryByBool(ctx context.Context, index string, must, should, mustNot []elastic.Query, result interface{}) error {
	query := elastic.NewBoolQuery()
	if must != nil {
		query.Must(must...)
	}
	if should != nil {
		query.Should(should...)
	}
	if mustNot != nil {
		query.MustNot(mustNot...)
	}
	searchResult, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	return parseSearchResult(searchResult, result)
}

// QueryByPhrase 根据短语查询文档
func (e *esClient) QueryByPhrase(ctx context.Context, index, field, phrase string, result interface{}) error {
	query := elastic.NewMatchPhraseQuery(field, phrase)
	searchResult, err := e.client.Search().Index(index).Query(query).Do(ctx)
	if err != nil {
		return err
	}
	return parseSearchResult(searchResult, result)
}

// parseSearchResult 解析搜索结果并将其解码到 value 中
func parseSearchResult(result *elastic.SearchResult, value interface{}) error {
	if result.Hits.TotalHits.Value > 0 {
		hits := make([]json.RawMessage, len(result.Hits.Hits))
		for i, hit := range result.Hits.Hits {
			hits[i] = hit.Source
		}
		// 使用 json.Marshal 将 json.RawMessage 切片转换为 []byte
		var buf bytes.Buffer
		buf.Write([]byte{'['})
		for i, hit := range hits {
			if i > 0 {
				buf.Write([]byte{','})
			}
			buf.Write(hit)
		}
		buf.Write([]byte{']'})
		return json.Unmarshal(buf.Bytes(), value)
	}
	return nil
}
