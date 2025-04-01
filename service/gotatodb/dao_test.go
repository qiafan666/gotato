package gotatodb

import (
	"context"
	"github.com/qiafan666/gotato"
	"github.com/qiafan666/gotato/commons/gcommon"
	"gorm.io/gorm"
)

type IDao interface {
	Tx() IDao
	Rollback()
	Commit() error
	Db() *gorm.DB
	WithContext(ctx context.Context) IDao
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

type imp struct {
	db           *gorm.DB
	defaultWhere map[string]interface{}
}

func New() IDao {
	//默认is_deleted=0条件
	defaultWhere := map[string]interface{}{}
	return &imp{
		db:           gotato.GetGotato().FeatureDB("test").GormDB(),
		defaultWhere: defaultWhere,
	}
}

func (i imp) Db() *gorm.DB {
	return i.db
}
func (i imp) Create(input interface{}) error {
	return i.db.Create(input).Error
}
func (i imp) Save(input interface{}) error {
	return i.db.Save(input).Error
}
func (i imp) First(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {

	if len(i.defaultWhere) > 0 {
		where = gcommon.MapMergeUnique(where, i.defaultWhere)
	}
	if scope != nil {
		i.db = i.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		i.db = i.db.Select(selectStr)
	} else {
		i.db = i.db.Select("*")
	}

	return i.db.Model(output).Where(where).First(output).Error
}
func (i imp) Find(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {

	if len(i.defaultWhere) > 0 {
		where = gcommon.MapMergeUnique(where, i.defaultWhere)
	}
	if scope != nil {
		i.db = i.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		i.db = i.db.Select(selectStr)
	} else {
		i.db = i.db.Select("*")
	}

	return i.db.Model(output).Where(where).Find(output).Error
}
func (i imp) Update(info interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if len(i.defaultWhere) > 0 {
		where = gcommon.MapMergeUnique(where, i.defaultWhere)
	}
	if scope != nil {
		i.db = i.db.Scopes(scope)
	}
	updateTx := i.db.Model(info).Where(where).Updates(info)
	err = updateTx.Error
	rows = updateTx.RowsAffected
	return
}

// UpdateMap 更新map结构体
// info 要更新的结构体
// table 要更新的表名
// where 更新条件
// scope 事务作用域
// jumpStrings 跳过结构体中的字段名
func (i imp) UpdateMap(info map[string]interface{}, table string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if len(i.defaultWhere) > 0 {
		where = gcommon.MapMergeUnique(where, i.defaultWhere)
	}
	if scope != nil {
		i.db = i.db.Scopes(scope)
	}

	updateTx := i.db.Table(table).Where(where).Updates(info)

	err = updateTx.Error
	rows = updateTx.RowsAffected
	return
}
func (i imp) Count(entity interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (total int64, err error) {

	if len(i.defaultWhere) > 0 {
		where = gcommon.MapMergeUnique(where, i.defaultWhere)
	}
	if scope != nil {
		i.db = i.db.Scopes(scope)
	}
	err = i.db.Model(entity).Where(where).Count(&total).Error
	return
}
func (i imp) Delete(entity interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {

	if scope != nil {
		i.db = i.db.Scopes(scope)
	}
	deleteTx := i.db.Model(entity).Where(where).Delete(&entity)
	rows = deleteTx.RowsAffected
	err = deleteTx.Error
	return
}
func (i imp) Tx() IDao {
	i.db = i.db.Begin()
	return IDao(&i)
}
func (i imp) WithContext(ctx context.Context) IDao {
	i.db = i.db.WithContext(ctx)
	return IDao(&i)
}
func (i imp) Rollback() {
	i.db.Rollback()
}
func (i imp) Commit() error {
	return i.db.Commit().Error
}
func (i imp) Raw(sql string, output interface{}) error {
	return i.db.Raw(sql).Scan(output).Error
}
func (i imp) DB() *gorm.DB {
	return i.db
}
