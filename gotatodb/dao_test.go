package gotatodb

import (
	"context"
	gotato "github.com/qiafan666/gotato"
	"github.com/qiafan666/gotato/commons"
	"gorm.io/gorm"
	"sync"
)

type Dao interface {
	Tx() Dao
	Rollback()
	Commit() error
	WithContext(ctx context.Context) Dao
	Create(interface{}) error
	First([]string, map[string]interface{}, func(*gorm.DB) *gorm.DB, interface{}) error
	Find([]string, map[string]interface{}, func(*gorm.DB) *gorm.DB, interface{}) error
	Update(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Delete(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Count(interface{}, map[string]interface{}, func(*gorm.DB) *gorm.DB) (int64, error)
	Save(interface{}) error
	Raw(string, interface{}) error
}

type Imp struct {
	db *gorm.DB
}

func (s Imp) Create(input interface{}) error {
	return s.db.Create(input).Error
}
func (s Imp) Save(input interface{}) error {
	return s.db.Save(input).Error
}
func (s Imp) First(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {
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
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	if value, ok := info.(map[string]interface{}); ok {
		table := value[commons.Table].(string)
		delete(value, commons.Table)
		dbs := s.db.Table(table).Where(where).Updates(info)
		err = dbs.Error
		rows = dbs.RowsAffected
	} else {
		dbs := s.db.Model(info).Where(where).Updates(info)
		err = dbs.Error
		rows = dbs.RowsAffected
	}
	return
}
func (s Imp) Count(entity interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (total int64, err error) {
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
	db := s.db.Model(entity).Where(where).Delete(&entity)
	rows = db.RowsAffected
	err = db.Error
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
	return &Imp{db: db}
}
