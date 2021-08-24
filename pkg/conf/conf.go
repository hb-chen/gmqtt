package conf

import (
	"github.com/hb-go/pkg/config"
	"github.com/hb-go/pkg/config/source"
	cliSource "github.com/hb-go/pkg/config/source/cli"
	"github.com/hb-go/pkg/config/source/file"
	"github.com/hb-go/pkg/log"
	"github.com/pborman/uuid"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

var (
	Conf Config // holds the global app config.
)

type Config struct {
	Log       Log
	Pyroscope Pyroscope

	logLevel string `json:"log_level"`

	SessionStore string `json:"session_store"`
	CacheStore   string `json:"cache_store"`

	// 应用配置
	App app

	Server server

	Auth auth

	Broker broker

	Sessions sessions

	// MySQL、PostgreSQL
	DB database `json:"database"`

	// Redis
	Redis redis

	// Opentracing
	Opentracing opentracing
}

type Log struct {
	Level string
	Debug bool
	E     bool
}

type Pyroscope struct {
	Enable bool
}

type app struct {
	Name      string `json:"name"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type server struct {
	Id     string
	Addr   string `json:"addr"`
	WsAddr string `json:"ws_addr"`
}

type auth struct {
	Provider string   `json:"provider"` // MockSuccess、MockFailure、rpc
	Addrs    []string `json:"addrs"`
}

type broker struct {
	Provider string   `json:"provider"` // mock、kafka
	Addrs    []string `json:"addrs"`
}

type sessions struct {
	Provider string `json:"provider"` // mock、redis
}

type database struct {
	Name     string `json:"name"`
	UserName string `json:"user_name"`
	Pwd      string `json:"pwd"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

type redis struct {
	Server string `json:"server"`
	Pwd    string `json:"pwd"`
}

type memcached struct {
	Server string `json:"server"`
}

type opentracing struct {
	Disable     bool   `json:"disable"`
	Type        string `json:"type"`
	ServiceName string `json:"service_name"`
	Address     string `json:"address"`
}

func init() {
}

// initConfig initializes the app configuration by first setting defaults,
// then overriding settings from the app config file, then overriding
// It returns an error if any.
func InitConfig(ctx *cli.Context) error {
	c, err := config.NewConfig()
	if err != nil {
		return err
	}

	sources := make([]source.Source, 0)

	// 1.files source
	patterns := ctx.StringSlice("config_patterns")
	for _, p := range patterns {
		files, err := filepath.Glob(p)
		if err != nil {
			return err
		}

		for _, f := range files {
			fileInfo, err := os.Lstat(f)
			if err != nil {
				return err
			}

			if fileInfo.IsDir() {
				log.Debugf("skipping directory: %s", f)
				continue
			}

			sources = append(sources, file.NewSource(file.WithPath(f)))
		}
	}

	// 2.cli source
	sources = append(sources, cliSource.WithContext(ctx))
	c.Load(
		sources...,
	)

	// Set defaults.
	Conf = Config{
		logLevel: "DEBUG",
	}

	c.Scan(&Conf)

	// @TODO 实例ID
	if len(Conf.Server.Id) == 0 {
		Conf.Server.Id = uuid.NewUUID().String()
	}

	// @TODO 配置检查
	log.Infof("config data:%v", Conf)

	return nil
}

func (c Config) LogLvl() log.Lvl {
	//DEBUG INFO WARN ERROR OFF
	switch c.logLevel {
	case "DEBUG":
		return log.DEBUG
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "OFF":
		return log.OFF
	default:
		return log.INFO
	}
}
