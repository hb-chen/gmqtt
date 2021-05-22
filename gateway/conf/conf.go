package conf

import (
	"github.com/hb-go/pkg/config"
	"github.com/hb-go/pkg/config/source"
	cliSource "github.com/hb-go/pkg/config/source/cli"
	"github.com/hb-go/pkg/config/source/file"
	"github.com/hb-go/pkg/log"
	"github.com/pborman/uuid"
	"github.com/urfave/cli/v2"
)

var (
	Conf              Config // holds the global app config.
	defaultConfigFile = "conf/conf.toml"
)

type Config struct {
	Log       Log
	Pyroscope Pyroscope

	logLevel string `toml:"log_level"`

	SessionStore string `toml:"session_store"`
	CacheStore   string `toml:"cache_store"`

	// 应用配置
	App app

	Server server

	Auth auth

	Broker broker

	Sessions sessions

	// MySQL、PostgreSQL
	DB database `toml:"database"`

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
	Name      string `toml:"name"`
	AccessKey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
}

type server struct {
	Id     string
	Addr   string `toml:"addr"`
	WsAddr string `toml:"ws_addr"`
}

type auth struct {
	Provider string   `toml:"provider"` // MockSuccess、MockFailure、rpc
	Addrs    []string `toml:"addrs"`
}

type broker struct {
	Provider string   `toml:"provider"` // mock、kafka
	Addrs    []string `toml:"addrs"`
}

type sessions struct {
	Provider string `toml:"provider"` // mock、redis
}

type database struct {
	Name     string `toml:"name"`
	UserName string `toml:"user_name"`
	Pwd      string `toml:"pwd"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
}

type redis struct {
	Server string `toml:"server"`
	Pwd    string `toml:"pwd"`
}

type memcached struct {
	Server string `toml:"server"`
}

type opentracing struct {
	Disable     bool   `toml:"disable"`
	Type        string `toml:"type"`
	ServiceName string `toml:"service_name"`
	Address     string `toml:"address"`
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
	// 1.file source
	if configFile := ctx.String("config_path"); len(configFile) > 0 {
		sources = append(sources, file.NewSource(file.WithPath(configFile)))
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
