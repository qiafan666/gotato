package gotatodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/glebarez/sqlite"
	slog "github.com/qiafan666/gotato/commons/log"
	serveries "github.com/qiafan666/gotato/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type GotatoDB struct {
	db     *gorm.DB
	name   string //db name
	dbType string //db 类型 mysql sqlite
}

func (slf *GotatoDB) GormDB() *gorm.DB {
	return slf.db
}

func (slf *GotatoDB) Name() string {
	return slf.name
}
func (slf *GotatoDB) StartPgsql(dbConfig serveries.DataBaseConfig) (err error) {
	if slf.db != nil {
		return errors.New("db already open")
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s", dbConfig.Addr, dbConfig.Username, dbConfig.Password, dbConfig.DbName, dbConfig.Port, dbConfig.Loc)
	slf.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{
		SingularTable: true,
	}, Logger: &slog.Gorm})
	if err != nil {
		slog.Slog.InfoF(context.Background(), "conn database error %s", err)
		return err
	}
	slf.name = dbConfig.Name
	db, err := slf.db.DB()
	if err != nil {
		slog.Slog.InfoF(context.Background(), "conn slf.db.DB() error %s", err)
		return err
	}
	db.SetConnMaxLifetime(dbConfig.MaxLifeTime * time.Millisecond)
	db.SetConnMaxIdleTime(dbConfig.MaxIdleTime * time.Millisecond)
	db.SetMaxOpenConns(dbConfig.MaxConn)
	db.SetMaxIdleConns(dbConfig.IdleConn)
	return nil
}
func (slf *GotatoDB) StartSqlite(dbConfig serveries.DataBaseConfig) error {
	if slf.db != nil {
		return errors.New("db already open")
	}
	slf.name = dbConfig.Name
	var err error
	slf.db, err = gorm.Open(sqlite.Open(dbConfig.DBFilePath), &gorm.Config{PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}, Logger: &slog.Gorm})
	if err != nil {
		slog.Slog.InfoF(context.Background(), "conn database error %s", err)
		return err
	}
	return nil
}
func (slf *GotatoDB) StartMysql(dbConfig serveries.DataBaseConfig) error {
	if slf.db != nil {
		return errors.New("db already open")
	}
	slf.name = dbConfig.Name
	var err error

	if len(dbConfig.Loc) <= 0 {
		dbConfig.Loc = "Local"
	}

	Dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Addr,
		dbConfig.Port,
		dbConfig.DbName,
		dbConfig.Charset,
		dbConfig.Loc,
	)
	slf.db, err = gorm.Open(mysql.Open(Dsn), &gorm.Config{PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		}, Logger: &slog.Gorm})
	if err != nil {
		slog.Slog.InfoF(context.Background(), "conn database error %s", err)
		return err
	}
	db, err := slf.db.DB()
	if err != nil {
		slog.Slog.InfoF(context.Background(), "conn slf.db.DB() error %s", err)
		return err
	}
	db.SetConnMaxLifetime(dbConfig.MaxLifeTime * time.Millisecond)
	db.SetConnMaxIdleTime(dbConfig.MaxIdleTime * time.Millisecond)
	db.SetMaxOpenConns(dbConfig.MaxConn)
	db.SetMaxIdleConns(dbConfig.IdleConn)
	return nil
}

func (slf *GotatoDB) StopDb() error {
	if slf.db != nil {
		db, err := slf.db.DB()
		if err != nil {
			slf.db = nil
		} else {
			_ = db.Close()
		}
		return err
	} else {
		return errors.New("db is nil")
	}
}
