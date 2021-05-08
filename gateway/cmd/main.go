package main

import (
	"context"
	"flag"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"strings"
	"sync"

	"net/url"

	"github.com/pyroscope-io/pyroscope/pkg/agent/profiler"

	"github.com/hb-chen/gmqtt/gateway/auth"
	"github.com/hb-chen/gmqtt/gateway/conf"
	"github.com/hb-chen/gmqtt/gateway/service"
	"github.com/hb-chen/gmqtt/pkg/log"
)

const (
	logCallerSkip = 2
)

var (
	cmdHelp      = flag.Bool("h", false, "帮助")
	confFilePath = flag.String("conf", "conf/conf.toml", "配置文件路径")
	//addr         = flag.String("addr", "tcp://127.0.0.1:1883", "server address")
	//etcdAddrs    = flag.String("etcdAddrs", "127.0.0.1:2379", "etcd address")
)

func init() {
	flag.Parse()

	l, err := initLogger("./log", "DEBUG", true, true)
	if err != nil {
		panic(err)
	}

	profiler.Start(profiler.Config{
		ApplicationName: "com.hbchen.gmqtt",

		// replace this with the address of pyroscope server
		ServerAddress: "http://pyroscope.pyroscope.svc.cluster.local:4040",

		// by default all profilers are enabled,
		//   but you can select the ones you want to use:
		ProfileTypes: []profiler.ProfileType{
			profiler.ProfileCPU,
			profiler.ProfileAllocObjects,
			profiler.ProfileAllocSpace,
			profiler.ProfileInuseObjects,
			profiler.ProfileInuseSpace,
		},

		Logger: l.Sugar(),
	})
}

func initLogger(path, level string, debug, e bool) (*zap.Logger, error) {
	logLevel := zapcore.WarnLevel
	err := logLevel.UnmarshalText([]byte(level))
	if err != nil {
		return nil, err
	}

	writer := logWriter(path)
	if e {
		stderr, close, err := zap.Open("stderr")
		if err != nil {
			close()
			return nil, err
		}
		writer = stderr
	}

	encoder := logEncoder(debug)
	core := zapcore.NewCore(encoder, writer, logLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(logCallerSkip))

	return logger, nil
}

func logEncoder(debug bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func logWriter(path string) zapcore.WriteSyncer {
	path = strings.TrimRight(path, "/")
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path + "/gmqtt.log",
		MaxSize:    10,
		MaxBackups: 10,
		MaxAge:     7,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func run(ctx *cli.Context) error {
	confPath := ctx.String("conf")

	// 配置初始化
	if err := conf.InitConfig(confPath); err != nil {
		return err
	}

	u, err := url.Parse(conf.Conf.Server.Addr)
	if err != nil {
		return err
	}

	log.SetColor(true)
	log.SetLevel(conf.Conf.LogLvl())

	if conf.Conf.Auth.Provider == auth.ProviderRpc {
		closer := auth.NewRpcRegister(conf.Conf.App.AccessKey, conf.Conf.App.SecretKey, conf.Conf.Auth.Addrs)
		defer func() {
			if err := closer.Close(); err != nil {
				log.Warnf("rpc auth close error:%v", err)
			}
		}()
	}

	wg := &sync.WaitGroup{}
	var wsServer *http.Server
	if len(conf.Conf.Server.WsAddr) > 0 {
		handler, err := service.WebsocketHandler("/mqtt", conf.Conf.Server.Addr)
		if err != nil {
			return err
		}

		wg.Add(1)
		wsServer = &http.Server{Addr: conf.Conf.Server.WsAddr, Handler: handler}
		go func() {
			defer wg.Done()
			if err := wsServer.ListenAndServe(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	server, err := service.NewServer()
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = server.ListenAndServe(u.Scheme, u.Host); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	if wsServer != nil {
		wsServer.Shutdown(ctx.Context)
	}
	server.Close()
	wg.Wait()

	return nil
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "conf",
			EnvVars: []string{"GM_CONF"},
			Usage:   "config file path.",
			Value:   "conf/conf.toml",
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = func(ctx *cli.Context) error {

		return run(ctx)
	}

	app.Commands = cli.Commands{
		&cli.Command{
			Name:  "reload",
			Usage: "TODO",
			Action: func(ctx *cli.Context) error {
				return nil
			},
		},
	}

	ctx := context.Background()
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
