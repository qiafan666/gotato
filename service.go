package gotato

import (
	"context"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/qiafan666/gotato/mongo"
	"github.com/qiafan666/gotato/oss"
	"os"
	"os/signal"
	"syscall"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	irisV12 "github.com/kataras/iris/v12"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/gconfig"
	"github.com/qiafan666/gotato/gotatodb"
	"github.com/qiafan666/gotato/iris"
	"github.com/qiafan666/gotato/redis"
)

// Instance we need create the single object but thread safe
var Instance *Server

type Server struct {
	app   iris.App
	redis []redis.Redis
	db    []gotatodb.GotatoDB
	oss   []oss.Oss
	mongo []mongo.Mongo
}
type ServerOption int

const (
	DatabaseService = iota + 1
	RedisService
	OssService
	MongoService
)

func init() {
	Instance = &Server{}
}

// GetGotatoInstance create the single object
func GetGotatoInstance() *Server {
	return Instance
}

func (slf *Server) Default() {
	slf.app.Default()
}

func (slf *Server) RegisterController(f func(app *irisV12.Application)) {
	f(slf.app.GetIrisApp())
}

func (slf *Server) RegisterErrorCodeAndMsg(language string, arr map[commons.ResponseCode]string) {
	if len(arr) == 0 {
		return
	}
	for k, v := range arr {
		commons.CodeMsg[language][k] = v
	}
}

func (slf *Server) WaitClose(params ...irisV12.Configurator) {
	defer glog.ZapLog.Sync()
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			// kill -SIGINT XXXX 或 Ctrl+c
			os.Interrupt,
			syscall.SIGINT, // register that too, it should be ok
			// os.Kill等同于syscall.Kill
			os.Kill,
			syscall.SIGKILL, // register that too, it should be ok
			// kill -SIGTERM XXXXD
			//^
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			glog.Slog.InfoF(context.Background(), "wait for close server")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for _, db := range slf.db {
				db.StopDb()
			}
			slf.app.GetIrisApp().Shutdown(ctx)
		}
	}()
	err := slf.app.Start(params...)
	if err != nil {
		panic(err)
	}
}

func (slf *Server) New() {
	slf.app.New()
}

// App return app
func (slf *Server) App() *iris.App {
	return &slf.app
}
func (slf *Server) FeatureDB(name string) *gotatodb.GotatoDB {
	for _, v := range slf.db {
		if v.Name() == name {
			return &v
		}
	}
	return nil
}
func (slf *Server) Redis(name string) *redisv8.Client {
	for _, v := range slf.redis {
		if v.Name() == name {
			return v.Redis()
		}
	}
	return nil
}
func (slf *Server) Mongo(name string) *mongo.Mongo {
	for _, v := range slf.mongo {
		if v.Name() == name {
			return &v
		}
	}
	return &mongo.Mongo{}
}
func (slf *Server) OssClient(name string) *alioss.Client {
	for _, v := range slf.oss {
		if v.Name() == name {
			return v.Client()
		}
	}
	return nil
}

func (slf *Server) OssBucket(name string) *alioss.Bucket {
	for _, v := range slf.oss {
		if v.Name() == name {
			return v.Bucket()
		}
	}
	return nil
}

func (slf *Server) LoadCustomizeConfig(slfConfig interface{}) {
	err := gconfig.LoadCustomizeConfig(slfConfig)
	if err != nil {
		panic(err)
	}
}

// StartServer need call this function after Option, if Dependent service is not started return panic.
func (slf *Server) StartServer(opt ...ServerOption) {
	var err error
	for _, v := range opt {
		switch v {
		case DatabaseService:
			slf.db = make([]gotatodb.GotatoDB, 0)
			for _, v := range gconfig.Configs.DataBase {
				if v.Type == "sqlite" {
					db := gotatodb.GotatoDB{}
					err = db.StartSqlite(v)
					if err != nil {
						panic(err)
					}
					slf.db = append(slf.db, db)
				} else if v.Type == "mysql" {
					db := gotatodb.GotatoDB{}
					err = db.StartMysql(v)
					if err != nil {
						panic(err)
					}
					slf.db = append(slf.db, db)
				} else if v.Type == "pgsql" {
					db := gotatodb.GotatoDB{}
					err = db.StartPgsql(v)
					if err != nil {
						panic(err)
					}
					slf.db = append(slf.db, db)
				} else {
					continue
				}
			}
		case RedisService:
			slf.redis = make([]redis.Redis, len(gconfig.Configs.Redis))
			for i, v := range gconfig.Configs.Redis {
				err = slf.redis[i].StartRedis(v)
				if err != nil {
					panic(err)
				}
			}
		case OssService:
			slf.oss = make([]oss.Oss, len(gconfig.Configs.Oss))
			for i, v := range gconfig.Configs.Oss {
				err = slf.oss[i].StartOss(v)
				if err != nil {
					panic(err)
				}
			}
		case MongoService:
			slf.mongo = make([]mongo.Mongo, len(gconfig.Configs.Mongo))
			for i, mongoConfig := range gconfig.Configs.Mongo {
				err = slf.mongo[i].StartMongo(mongoConfig)
				if err != nil {
					panic(err)
				}
			}
		}

	}
}
