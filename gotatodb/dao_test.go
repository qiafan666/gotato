package gotatodb

import (
	"context"
	gotato "github.com/qiafan666/gotato"
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
}

type Imp struct {
	db *gorm.DB
}

func (s Imp) Create(input interface{}) error {
	return s.db.Create(input).Error
}
func (s Imp) First(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		s.db = s.db.Select(selectStr)
	}

	return s.db.Model(output).Where(where).First(output).Error
}
func (s Imp) Find(selectStr []string, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB, output interface{}) (err error) {
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	if len(selectStr) > 0 {
		s.db = s.db.Select(selectStr)
	}

	return s.db.Model(output).Where(where).Find(output).Error
}
func (s Imp) Update(info interface{}, where map[string]interface{}, scope func(*gorm.DB) *gorm.DB) (rows int64, err error) {
	if scope != nil {
		s.db = s.db.Scopes(scope)
	}
	db := s.db.Model(info).Where(where).Updates(info)
	err = db.Error
	rows = db.RowsAffected
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

var db *gorm.DB
var once sync.Once

func Instance() Dao {
	once.Do(func() {
		db = gotato.GetGotatoInstance().FeatureDB("metaspace").GormDB()
	})
	return &Imp{db: db}
}
