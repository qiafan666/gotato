package v2

import (
	"context"
	"errors"
	"fmt"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/qiafan666/gotato/commons"
	slog "github.com/qiafan666/gotato/commons/log"
	"github.com/qiafan666/gotato/config"
	"github.com/qiafan666/gotato/gotatodb"
	"github.com/qiafan666/gotato/mongo"
	"github.com/qiafan666/gotato/oss"
	"github.com/qiafan666/gotato/redis"
	"github.com/qiafan666/gotato/v2/middleware"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	redisV8 "github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Instance we need create the single object but thread safe
var Instance *Server

type Server struct {
	app        *gin.Engine
	redis      []redis.Redis
	db         []gotatodb.GotatoDB
	ctx        context.Context
	httpServer *http.Server
	oss        []oss.Oss
	mongo      []mongo.Mongo
}
type ServerOption int

const (
	DatabaseService = iota + 1
	RedisService
	GinService
	GinInitService
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
func (slf *Server) RegisterErrorCodeAndMsg(language string, arr map[commons.ResponseCode]string) {
	commons.RegisterCodeAndMsg(language, arr)
}

func (slf *Server) WaitClose() {
	defer func(ZapLog *zap.SugaredLogger) {
		_ = ZapLog.Sync()
	}(slog.ZapLog)
	//创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.SC.SConfigure.Port),
		Handler: slf.app,
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		// kill -SIGINT XXL 或 Ctrl+c
		os.Interrupt,
		syscall.SIGINT, // register that too, it should be ok
		// os.Kill等同于syscall.Kill
		os.Kill,
		syscall.SIGKILL, // register that too, it should be ok
		// kill -SIGTERM XXE
		//^
		syscall.SIGTERM,
	)
	select {
	case <-ch:
		slog.Slog.InfoF(context.Background(), "wait for close server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, db := range slf.db {
			_ = db.StopDb()
		}
		for _, stopRedis := range slf.redis {
			_ = stopRedis.StopRedis()
		}
		err := server.Shutdown(ctx)
		if err != nil {
			slog.Slog.ErrorF(context.Background(), err.Error())
		}
	}
}

// App return app
func (slf *Server) App() *gin.Engine {
	return slf.app
}
func (slf *Server) FeatureDB(name string) *gotatodb.GotatoDB {
	for _, v := range slf.db {
		if v.Name() == name {
			return &v
		}
	}
	return nil
}
func (slf *Server) Redis(name string) *redisV8.Client {
	for _, v := range slf.redis {
		if v.Name() == name {
			return v.Redis()
		}
	}
	return nil
}
func (slf *Server) Mongo(name string) mongo.Mongo {
	for _, v := range slf.mongo {
		if v.Name() == name {
			return v
		}
	}
	return mongo.Mongo{}
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
	err := config.LoadCustomizeConfig(slfConfig)
	if err != nil {
		panic(err)
	}
}
func (slf *Server) gin() {
	//设置模式
	if config.SC.SConfigure.Profile == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.SC.SConfigure.Profile == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	gin.ForceConsoleColor()
	slf.app = gin.New()

	slf.app.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, commons.BuildFailed(commons.HttpNotFound, commons.DefaultLanguage, ""))
		ctx.Abort()
		return
	})
	slf.app.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, commons.BuildFailed(commons.HttpNotFound, commons.DefaultLanguage, ""))
		ctx.Abort()
		return
	})

	//插入中间件
	slf.app.Use(middleware.Default)

	slf.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.SC.SConfigure.Port),
		Handler: slf.App(),
	}
	//开启pprof
	if config.SC.PProfConfig.Enable == true {
		slf.app.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	}

	//开启swagger
	if config.SC.SwaggerConfig.Enable == true {
		slf.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	//忽略pprof和swagger的路由日志
	middleware.RegisterIgnoreRequest("/debug/pprof/*", "/swagger/*")

	go func() {
		if err := slf.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Slog.ErrorF(context.Background(), err.Error())
		}
	}()
}
func (slf *Server) ginInit() {

	//设置模式
	if config.SC.SConfigure.Profile == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.SC.SConfigure.Profile == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	slf.app = gin.New()

	slf.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.SC.SConfigure.Port),
		Handler: slf.App(),
	}
	//开启pprof
	if config.SC.PProfConfig.Enable == true {
		slf.app.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	}

	//开启swagger
	if config.SC.SwaggerConfig.Enable == true {
		slf.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	go func() {
		if err := slf.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Slog.ErrorF(context.Background(), err.Error())
		}
	}()
}

// StartServer need call this function after Option, if Dependent service is not started return panic.
func (slf *Server) StartServer(opt ...ServerOption) {
	var err error
	for _, v := range opt {
		switch v {
		case GinService:
			slf.gin()
		case GinInitService:
			slf.ginInit()
		case DatabaseService:
			slf.db = make([]gotatodb.GotatoDB, 0)
			for _, v := range config.Configs.DataBase {
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
			slf.redis = make([]redis.Redis, len(config.Configs.Redis))
			for i, v := range config.Configs.Redis {
				err = slf.redis[i].StartRedis(v)
				if err != nil {
					panic(err)
				}
			}
		case OssService:
			slf.oss = make([]oss.Oss, len(config.Configs.Oss))
			for i, v := range config.Configs.Oss {
				err = slf.oss[i].StartOss(v)
				if err != nil {
					panic(err)
				}
			}
		case MongoService:
			slf.mongo = make([]mongo.Mongo, len(config.Configs.Mongo))
			for i, mongoConfig := range config.Configs.Mongo {
				err = slf.mongo[i].StartMongo(mongoConfig)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
