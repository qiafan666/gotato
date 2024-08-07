package mongo

import (
	"github.com/qiafan666/gotato"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type Dao interface {
	Insert(collection string, docs any) (err error)
	InsertManyData(collection string, docs []any) (err error)
	IsExistKey(collection string, bs bson.M) (flag bool, err error)
	QueryData(collection string, bs bson.M, value any) (err error)
	GetCount(collection string, bs bson.M) (count int, err error)
	QueryAllData(collection string, bs bson.M, value any) (err error)
	QueryMultiData(collection string, bs bson.M, sort string, offset, limit int, value any) (count int, err error)
	UpdateData(collection string, bs bson.M, value any) (err error)
	UpdateAllData(collection string, bs bson.M, value any) (count int, err error)
	UpdateDataById(collection string, id, value any) (err error)
	Upsert(collection string, bs bson.M, value any) (err error)
	DelDataById(collection string, id any) (err error)
	DelData(collection string, bs bson.M) error
	DelAllData(collection string, bs bson.M) (int, error)
	Aggregate(collection string, bs []bson.M, value any) error
	AggregateOne(collection string, bs []bson.M, value any) error
	CreateIndex(collection string, unique bool, timeLimit bool, value []string) error
	CloseMongoDb()
}

type Imp struct {
	mongo *Mongo
}

func Instance() Dao {
	return &Imp{mongo: gotato.GetGotatoInstance().Mongo("test")}
}

// Insert 向指定集合插入单个文档
func (i *Imp) Insert(collection string, docs any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Insert(docs)
}

// InsertManyData 向指定集合插入多个文档
func (i *Imp) InsertManyData(collection string, docs []any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Insert(docs...)
}

// IsExistKey 检查集合中是否存在符合条件的文档
func (i *Imp) IsExistKey(collection string, bs bson.M) (bool, error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	var (
		value any
	)
	if err := s.DB(i.mongo.DB).C(collection).Find(bs).One(value); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// QueryData 查询集合中符合条件的单个文档
func (i *Imp) QueryData(collection string, bs bson.M, value any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	err = s.DB(i.mongo.DB).C(collection).Find(bs).One(value)
	return
}

// GetCount 获取符合条件的文档数量
func (i *Imp) GetCount(collection string, bs bson.M) (count int, err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Find(bs).Count()
}

// QueryAllData 查询集合中所有符合条件的文档
func (i *Imp) QueryAllData(collection string, bs bson.M, value any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Find(bs).All(value)
}

// QueryMultiData 查询集合中符合条件的多个文档并排序、分页
func (i *Imp) QueryMultiData(collection string, bs bson.M, sort string, offset, limit int, value any) (int, error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	count, err := s.DB(i.mongo.DB).C(collection).Find(bs).Count()
	if err != nil {
		return 0, err
	}
	if err = s.DB(i.mongo.DB).C(collection).Find(bs).Sort(sort).Skip(offset).Limit(limit).All(value); err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateData 更新符合条件的单个文档
func (i *Imp) UpdateData(collection string, bs bson.M, value any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Update(bs, value)
}

// UpdateAllData 更新所有符合条件的文档
func (i *Imp) UpdateAllData(collection string, bs bson.M, value any) (num int, err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	info, err := s.DB(i.mongo.DB).C(collection).UpdateAll(bs, value)
	if err != nil {
		return 0, err
	}
	return info.Updated, err
}

// UpdateDataById 根据ID更新单个文档
func (i *Imp) UpdateDataById(collection string, id, value any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).UpdateId(id, value)
}

// Upsert 插入或更新符合条件的文档
func (i *Imp) Upsert(collection string, bs bson.M, value any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	_, err = s.DB(i.mongo.DB).C(collection).Upsert(bs, value)
	return err
}

// DelDataById 根据ID删除单个文档
func (i *Imp) DelDataById(collection string, id any) (err error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).RemoveId(id)
}

// DelData 删除符合条件的单个文档
func (i *Imp) DelData(collection string, bs bson.M) error {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Remove(bs)
}

// DelAllData 删除所有符合条件的文档
func (i *Imp) DelAllData(collection string, bs bson.M) (int, error) {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	info, err := s.DB(i.mongo.DB).C(collection).RemoveAll(bs)
	if err != nil {
		return 0, err
	} else {
		return info.Removed, nil
	}
}

// Aggregate 进行聚合查询，返回多个结果
func (i *Imp) Aggregate(collection string, bs []bson.M, value any) error {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Pipe(bs).All(value)
}

// AggregateOne 进行聚合查询，返回单个结果
func (i *Imp) AggregateOne(collection string, bs []bson.M, value any) error {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	return s.DB(i.mongo.DB).C(collection).Pipe(bs).One(value)
}

// CreateIndex 为集合创建索引
func (i *Imp) CreateIndex(collection string, unique bool, timeLimit bool, value []string) error {
	s := i.mongo.c.Ref()
	defer i.mongo.c.UnRef(s)
	index := mgo.Index{
		Key:        value,
		Unique:     unique,
		Background: true,
		// ExpireAfter: 1 * time.Minute,
	}
	return s.DB(i.mongo.DB).C(collection).EnsureIndex(index)
}

func (i *Imp) CloseMongoDb() {
	i.mongo.c.Close()
}
