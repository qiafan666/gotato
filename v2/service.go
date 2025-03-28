package v2

import (
	"context"
	"errors"
	"fmt"
	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/qiafan666/gotato/commons"
	"github.com/qiafan666/gotato/commons/gerr"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/gconfig"
	"github.com/qiafan666/gotato/gotatodb"
	"github.com/qiafan666/gotato/oss"
	"github.com/qiafan666/gotato/redis"
	"github.com/qiafan666/gotato/v2/middleware"
	redisV9 "github.com/redis/go-redis/v9"

	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Instance we need create the single object but thread safe
var instance *Server

type Server struct {
	app        *gin.Engine
	redis      []redis.Redis
	db         []gotatodb.GotatoDB
	httpServer *http.Server
	oss        []oss.Oss
	ctx        context.Context
	cancel     context.CancelFunc
}
type ServerOption int

const (
	DatabaseService = iota + 1
	RedisService
	GinService
	GinInitService
	OssService
)

func init() {
	instance = &Server{}
}

func (slf *Server) GetCtx() context.Context {
	return slf.ctx
}

func (slf *Server) GetCancel() context.CancelFunc {
	return slf.cancel
}

func (slf *Server) SetMysqlLogCallerSkip(skip int) {
	glog.GormSkip = skip
	glog.ReInit()
}

// GetGotatoInstance create the single object
func GetGotatoInstance() *Server {
	return instance
}
func (slf *Server) RegisterErrorCodeAndMsg(language string, arr map[int]string) {
	gerr.RegisterCodeAndMsg(language, arr)
}

func (slf *Server) WaitClose(stopFunc ...func()) {
	defer func(ZapLog *zap.SugaredLogger) {
		_ = ZapLog.Sync()
	}(glog.ZapLog)
	//创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", gconfig.SC.SConfigure.Port),
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
		glog.Slog.InfoF(nil, "wait for close server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, db := range slf.db {
			_ = db.StopDb()
		}
		for _, stopRedis := range slf.redis {
			_ = stopRedis.StopRedis()
		}
		for {
			if atomic.LoadInt64(&commons.ActiveRequests) == 0 {
				break
			}
			time.Sleep(time.Second)
		}

		if gconfig.SC.FeiShuConfig.Enable {
			glog.FeiShu.Close()
		}

		//关闭HTTP服务器之前关闭传入的stopFunc
		if len(stopFunc) > 0 {
			for _, f := range stopFunc {
				f()
			}
		}

		//关闭主context
		slf.cancel()
		time.Sleep(time.Second)
		err := server.Shutdown(ctx)
		if err != nil {
			glog.Slog.ErrorF(nil, "server shutdown error: %s", err.Error())
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

func (slf *Server) Redis(name string) redisV9.UniversalClient {
	for _, v := range slf.redis {
		if v.Name() == name {
			return v.Redis()
		}
	}
	return nil
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
func (slf *Server) gin() {
	//设置模式
	if gconfig.SC.SConfigure.Profile == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if gconfig.SC.SConfigure.Profile == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	gin.ForceConsoleColor()
	slf.app = gin.New()

	slf.app.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, commons.BuildFailed(gerr.HttpNotFound, gerr.DefaultLanguage, ""))
		ctx.Abort()
		return
	})
	slf.app.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, commons.BuildFailed(gerr.HttpNotFound, gerr.DefaultLanguage, ""))
		ctx.Abort()
		return
	})

	//插入中间件
	slf.app.Use(middleware.Default)

	slf.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", gconfig.SC.SConfigure.Addr, gconfig.SC.SConfigure.Port),
		Handler: slf.App(),
	}
	//开启pprof
	if gconfig.SC.PProfConfig.Enable == true {
		slf.app.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	}

	//开启swagger
	if gconfig.SC.SwaggerConfig.Enable == true {
		slf.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	//忽略pprof和swagger的路由日志
	middleware.RegisterIgnoreRequest("/debug/pprof/*", "/swagger/*")

	go func() {
		if err := slf.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			glog.Slog.ErrorF(nil, "listen and serve error: %s", err.Error())
		}
	}()
}
func (slf *Server) ginInit() {

	//设置模式
	if gconfig.SC.SConfigure.Profile == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else if gconfig.SC.SConfigure.Profile == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	slf.app = gin.New()

	slf.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", gconfig.SC.SConfigure.Addr, gconfig.SC.SConfigure.Port),
		Handler: slf.App(),
	}
	//开启pprof
	if gconfig.SC.PProfConfig.Enable == true {
		slf.app.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	}

	//开启swagger
	if gconfig.SC.SwaggerConfig.Enable == true {
		slf.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	go func() {
		if err := slf.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			glog.Slog.ErrorF(nil, "listen and serve error: %s", err.Error())
		}
	}()
}

// StartServer need call this function after Option, if Dependent service is not started return panic.
func (slf *Server) StartServer(opt ...ServerOption) {
	slf.ctx, slf.cancel = context.WithCancel(context.Background())
	var err error
	for _, v := range opt {
		switch v {
		case GinService:
			slf.gin()
		case GinInitService:
			slf.ginInit()
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
		}
	}
}
