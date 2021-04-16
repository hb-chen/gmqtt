package conf

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/pborman/uuid"

	"github.com/hb-chen/gmqtt/pkg/log"
)

var (
	Conf              config // holds the global app config.
	defaultConfigFile = "conf/conf.toml"
)

type config struct {
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
func InitConfig(configFile string) error {
	if configFile == "" {
		configFile = defaultConfigFile
	}

	// Set defaults.
	Conf = config{
		logLevel: "DEBUG",
	}

	if _, err := os.Stat(configFile); err != nil {
		return errors.New("config file err:" + err.Error())
	} else {
		log.Infof("load config from file:" + configFile)
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return errors.New("config load err:" + err.Error())
		}
		_, err = toml.Decode(string(configBytes), &Conf)
		if err != nil {
			return errors.New("config decode err:" + err.Error())
		}
	}

	// @TODO 实例ID
	if len(Conf.Server.Id) == 0 {
		Conf.Server.Id = uuid.NewUUID().String()
	}

	// @TODO 配置检查
	log.Infof("config data:%v", Conf)

	return nil
}

func (c config) LogLvl() log.Lvl {
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
