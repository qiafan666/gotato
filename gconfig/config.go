package gconfig

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var SC ServerConfig
var Configs Config
var YamlFile []byte

func ReadConfig() {
	yamlFile, err := os.ReadFile("application.yaml")
	if err != nil {
		panic(fmt.Errorf("load application.yaml error, will exit,please fix the application"))
	}
	err = yaml.Unmarshal(yamlFile, &SC)
	if err != nil {
		panic(err)
	}
	for _, v := range os.Args {
		if strings.Contains(v, "-env") {
			SC.SConfigure.Profile = strings.Split(v, "=")[1]
		}
	}
	if len(SC.SConfigure.Profile) == 0 {
		// load dev profile application-dev.yaml
		Configs = InitAllConfig(filepath.Join(filepath.Clean(SC.SConfigure.ConfigPath), "dev"))
	} else {
		Configs = InitAllConfig(filepath.Join(filepath.Clean(SC.SConfigure.ConfigPath), SC.SConfigure.Profile))
	}
}

type DataBaseConfig struct {
	//common
	Name    string        `yaml:"name"`
	Type    string        `yaml:"type"`
	SlowSql time.Duration `yaml:"slow_sql"`

	//sqlite
	DBFilePath string `yaml:"db_file_path"`

	//mysql
	Addr        string        `yaml:"addr"`
	Port        string        `yaml:"port"`
	Username    string        `yaml:"username"`
	DbName      string        `yaml:"db_name"`
	Loc         string        `json:"loc"`
	Password    string        `yaml:"password"`
	IdleConn    int           `yaml:"idle_conn"`
	MaxConn     int           `yaml:"max_conn"`
	MaxIdleTime time.Duration `yaml:"max_idle_time"`
	MaxLifeTime time.Duration `yaml:"max_life_time"`
	Charset     string        `yaml:"charset"`
}

type RedisConfig struct {
	ClusterMode bool     `yaml:"cluster_mode"` // 是否启用集群模式。
	Address     []string `yaml:"address"`      // 地址
	Username    string   `yaml:"username"`     // 用户名
	Password    string   `yaml:"password"`     // 密码
	MaxRetry    int      `yaml:"max_retry"`    // 最大重试次数。
	DB          int      `yaml:"db"`           // 数据库编号。
	PoolSize    int      `yaml:"pool_size"`    // 连接池大小。
	Name        string   `yaml:"name"`         // 名称
}

type Config struct {
	DataBase []DataBaseConfig `yaml:"dataSource"`
	Redis    []RedisConfig    `yaml:"redis"`
	Oss      []OssConfig      `yaml:"oss"`
	Smtp     []SmtpConfig     `yaml:"smtp"`
}

type OssConfig struct {
	OssBucket       string `yaml:"oss_bucket"`
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	OssEndPoint     string `yaml:"oss_end_point"`
	Name            string `yaml:"name"`
}

type SmtpConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Sender   string `yaml:"sender"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

func InitAllConfig(fileName string) Config {
	dir, err := os.ReadDir(fileName)
	if err != nil {
		log.Panicf("load config error %s %s", err, fileName)
	}
	var buffer bytes.Buffer
	for _, v := range dir {
		if v.IsDir() == false {
			if strings.Contains(v.Name(), ".yaml") {
				file, err := os.ReadFile(fileName + "/" + v.Name())
				if err != nil {
					panic("load config error")
				}
				buffer.Write(file)
				buffer.Write([]byte("\n"))
				continue
			}
		}
	}
	dbc := Config{}
	YamlFile = buffer.Bytes()
	err = yaml.Unmarshal(buffer.Bytes(), &dbc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	return dbc
}

func LoadCustomizeConfig(config interface{}) error {
	err := yaml.Unmarshal(YamlFile, config)
	if err != nil {
		return err
	}
	return nil
}

type ServerBaseConfig struct {
	Addr         string `yaml:"addr"`
	Port         int    `yaml:"port"`
	GormLogLevel string `yaml:"gorm_log_level"`
	ZapLogLevel  string `yaml:"zap_log_level"`
	Profile      string `yaml:"profile"`
	LogPath      string `yaml:"logPath"`
	LogName      string `yaml:"logName"`
	ConfigPath   string `yaml:"configPath"`
	SimpleStdout bool   `yaml:"simple_stdout"`
	FunctionName bool   `yaml:"function_name"`
}
type ServerConfig struct {
	SConfigure    ServerBaseConfig `yaml:"server"`
	SwaggerConfig SwaggerConfig    `yaml:"swagger"`
	PProfConfig   PProfConfig      `yaml:"pprof"`
	FeiShuConfig  FeiShuConfig     `yaml:"feiShu"`
}
type SwaggerConfig struct {
	Enable   bool   `yaml:"enable"`
	UiPath   string `yaml:"ui_path"`
	JsonPath string `yaml:"json_path"`
}
type PProfConfig struct {
	Enable bool   `yaml:"enable"`
	Port   string `yaml:"port"`
}
type FeiShuConfig struct {
	Enable  bool   `yaml:"enable"`
	Url     string `yaml:"url"`
	GroupId string `yaml:"group_id"`
}
