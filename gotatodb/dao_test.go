package gotatodb

import (
	"context"
	gotato "github.com/qiafan666/gotato"
	"github.com/qiafan666/gotato/commons/gmap"
	"gorm.io/gorm"
	"sync"
)

type Dao interface {
	Tx() Dao
	Rollback()
	Commit() error
	Db() *gorm.DB
	WithContext(ctx context.Context) Dao
	Create(interface{}) error
	First([]string, map[string]interface{}, func(*gorm.DB) *gorm.DB, interface{}) error
	Find([]string, map[string]interface{}, func(*gorm.DB) *gorm.DB, interface{}) error
	Update(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	UpdateMap(map[string]interface{}, string, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Delete(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Count(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Save(interface{}) error
	Raw(string, interface{}) error
}

type Imp struct {
	db           *gorm.DB
	defaultWhere map[string]interface{}
}

func (s Imp) Db() *gorm.DB {
	return s.db
}
func (s Imp) Create(input interface{}) error {
	return s.db.Create(input).Error
}
func (s Imp) Save(input interface{}) error {
	return s.db.Save(input).Error
}
func (s Imp) First(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {

	if len(s.defaultWhere) > 0 {
		where = gmap.MergeMapsUnique(where, s.defaultWhere)
	}
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		s.db = s.db.Select(selectStr)
	} else {
		s.db = s.db.Select("*")
	}

	return s.db.Model(output).Where(where).First(output).Error
}
func (s Imp) Find(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {

	if len(s.defaultWhere) > 0 {
		where = gmap.MergeMapsUnique(where, s.defaultWhere)
	}
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		s.db = s.db.Select(selectStr)
	} else {
		s.db = s.db.Select("*")
	}

	return s.db.Model(output).Where(where).Find(output).Error
}
func (s Imp) Update(info interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if len(s.defaultWhere) > 0 {
		where = gmap.MergeMapsUnique(where, s.defaultWhere)
	}
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	updates := s.db.Model(info).Where(where).Updates(info)
	err = updates.Error
	rows = updates.RowsAffected
	return
}

// UpdateMap 更新map结构体
// info 要更新的结构体
// table 要更新的表名
// where 更新条件
// scope 事务作用域
// jumpStrings 跳过结构体中的字段名
func (s Imp) UpdateMap(info map[string]interface{}, table string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if len(s.defaultWhere) > 0 {
		where = gmap.MergeMapsUnique(where, s.defaultWhere)
	}
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}

	updates := s.db.Table(table).Where(where).Updates(info)

	err = updates.Error
	rows = updates.RowsAffected
	return
}
func (s Imp) Count(entity interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (total int64, err error) {

	if len(s.defaultWhere) > 0 {
		where = gmap.MergeMapsUnique(where, s.defaultWhere)
	}
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	err = s.db.Model(entity).Where(where).Count(&total).Error
	return
}
func (s Imp) Delete(entity interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	deletes := s.db.Model(entity).Where(where).Delete(&entity)
	rows = deletes.RowsAffected
	err = deletes.Error
	return
}
func (s Imp) Tx() Dao {
	s.db = s.db.Begin()
	return Dao(&s)
}
func (s Imp) WithContext(ctx context.Context) Dao {
	s.db = s.db.WithContext(ctx)
	return Dao(&s)
}
func (s Imp) Rollback() {
	s.db.Rollback()
}
func (s Imp) Commit() error {
	return s.db.Commit().Error
}
func (s Imp) Raw(sql string, output interface{}) error {
	return s.db.Raw(sql).Scan(output).Error
}
func (s Imp) DB() *gorm.DB {
	return s.db
}

var db *gorm.DB
var once sync.Once

func Instance() Dao {
	once.Do(func() {
		db = gotato.GetGotatoInstance().FeatureDB("test").GormDB()
	})

	//默认is_deleted=0条件
	defaultWhere := map[string]interface{}{}

	return &Imp{db: db, defaultWhere: defaultWhere}
}
