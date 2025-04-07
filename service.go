package gotato

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/qiafan666/gotato/commons/ggin"
	"github.com/qiafan666/gotato/commons/glog"
	"github.com/qiafan666/gotato/service/gconfig"
	"github.com/qiafan666/gotato/service/gotatodb"
	"github.com/qiafan666/gotato/service/oss"
	"github.com/qiafan666/gotato/service/redis"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/qiafan666/gotato/commons/gcommon"
	"github.com/qiafan666/gotato/commons/gerr"
	redisV9 "github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Instance we need create the single object but thread safe
var instance *Server

var ActiveRequests int64

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
	ctx, cancel := context.WithCancel(context.Background())
	instance = &Server{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (slf *Server) GetCtx() context.Context {
	return slf.ctx
}

func (slf *Server) GetCancel() context.CancelFunc {
	return slf.cancel
}

// GetGotato create the single object
func GetGotato() *Server {
	return instance
}

func (slf *Server) ReadConfig() {
	gconfig.ReadConfig()
	glog.NewZap()
}

func (slf *Server) SetMysqlLogCallerSkip(skip int) {
	glog.GormSkip = skip
	glog.ReNewZap()
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
			if atomic.LoadInt64(&ActiveRequests) == 0 {
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

func (slf *Server) LoadCustomCfg(slfConfig interface{}) {
	err := gconfig.LoadCfg(slfConfig)
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
		ggin.GinError(ctx, gerr.NewLang(gerr.HttpNotFound, gerr.DefaultLanguage, ""))
		ctx.Abort()
		return
	})
	slf.app.NoMethod(func(ctx *gin.Context) {
		ggin.GinError(ctx, gerr.NewLang(gerr.HttpNotFound, gerr.DefaultLanguage, ""))
		ctx.Abort()
		return
	})

	//插入中间件
	slf.app.Use(slf.trace)

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
	slf.RegisterIgnoreRequest("/debug/pprof/*", "/swagger/*")

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

// -------------------- middleware --------------------

var ignoreRequestMap sync.Map

// RegisterIgnoreRequest 忽略打印当前路径的接口日志,支持/*通配符
func (slf *Server) RegisterIgnoreRequest(paths ...string) {
	for _, path := range paths {
		// 如果路径中包含通配符 /*，则将其替换为正则表达式中的通配符 .*
		currentPath := path
		if strings.Contains(path, "/*") {
			currentPath = strings.Replace(path, "/*", "/.*", -1)
		}

		if _, exist := ignoreRequestMap.Load(currentPath); !exist {
			ignoreRequestMap.Store(currentPath, true)
		}
	}
}

// IsIgnoredRequest 判断请求路径是否应该被忽略
func (slf *Server) isIgnoredRequest(requestPath string) bool {
	var isIgnored bool
	ignoreRequestMap.Range(func(key, value interface{}) bool {
		pathPattern := key.(string)
		// 使用正则表达式匹配请求路径
		matched, err := regexp.MatchString(pathPattern, requestPath)
		if err == nil && matched {
			isIgnored = true
			return false // 停止 Range 循环
		}
		return true // 继续 Range 循环
	})
	return isIgnored
}

// cstomResponseWriter 自定义响应写入器结构体
type cstomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 实现了 io.Writer 接口中的 Write 方法，用于写入字节切片，并记录到响应体中
func (w *cstomResponseWriter) Write(b []byte) (int, error) {
	// 将字节切片写入到响应体中
	n, err := w.body.Write(b)
	if err != nil {
		return n, err
	}
	// 写入响应体
	return w.ResponseWriter.Write(b)
}

// WriteString 实现了 WriteString 方法，用于写入字符串，并记录到响应体中
func (w *cstomResponseWriter) WriteString(s string) (int, error) {
	// 将字符串写入到响应体中
	n, err := w.body.WriteString(s)
	if err != nil {
		return n, err
	}
	// 写入响应体
	return w.ResponseWriter.WriteString(s)
}

func (slf *Server) trace(ctx *gin.Context) {
	header := ctx.GetHeader("request_id")
	if header != "" {
		ctx.Set("request_id", header)
		ctx.Set("ctx", context.WithValue(ctx, "request_id", header))
	} else {
		uuid := gcommon.GenerateUUID()
		value := context.WithValue(ctx, "request_id", uuid)
		ctx.Set("request_id", uuid)
		ctx.Set("ctx", value)
	}

	atomic.AddInt64(&ActiveRequests, 1)
	defer atomic.AddInt64(&ActiveRequests, -1)
	defer func() {
		if err := recover(); err != nil {
			var stacktrace string
			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)
				if !got {
					break
				}
				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}
			// when stack finishes
			logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
			logMessage += fmt.Sprintf("Trace: %s", err)
			logMessage += fmt.Sprintf("\n%s", stacktrace)
			glog.Slog.ErrorF(ctx, logMessage)
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}()

	blw := &cstomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw

	if !slf.isIgnoredRequest(ctx.Request.URL.Path) {
		var bodyBytes []byte
		var err error
		var requestBody *bytes.Buffer
		if ctx.Request.Method == http.MethodPost {
			bodyBytes, err = io.ReadAll(ctx.Request.Body)
			if err != nil {
				glog.Slog.ErrorF(ctx, "ReadAll %s", err)
			} else if len(bodyBytes) > 0 {
				requestBody = bytes.NewBuffer(bodyBytes)
				ctx.Request.Body = io.NopCloser(requestBody)
			} else {
				requestBody = bytes.NewBuffer([]byte(""))
			}
		}
		start := time.Now()
		ctx.Next()

		path := ctx.Request.URL.Path
		if ctx.Request.URL.RawQuery != "" {
			path += "?" + ctx.Request.URL.RawQuery
		}

		if gconfig.SC.SConfigure.SimpleStdout {
			glog.Slog.InfoF(ctx, "[%s:%s][%s][%dms][response code:%d]",
				ctx.Request.Method, path, gcommon.RemoteIP(ctx.Request), time.Now().Sub(start).Milliseconds(), ctx.Writer.Status())
		} else {
			glog.Slog.InfoF(ctx, "[%s:%s][%s][%dms][response code:%d][request:%s][response:%s]",
				ctx.Request.Method, path, gcommon.RemoteIP(ctx.Request), time.Now().Sub(start).Milliseconds(),
				ctx.Writer.Status(), strings.ReplaceAll(strings.Replace(string(bodyBytes), "\n", "", -1), " ", ""), blw.body.String())
		}
	} else {
		ctx.Next()
	}
}
