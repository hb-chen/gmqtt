package main

import (
	"flag"
	"net/url"

	"github.com/hb-go/micro-mq/pkg/log"
	. "github.com/hb-go/micro-mq/gateway/conf"
	"github.com/hb-go/micro-mq/gateway/auth"
	"github.com/hb-go/micro-mq/gateway/service"
)

var (
	cmdHelp      = flag.Bool("h", false, "帮助")
	confFilePath = flag.String("conf", "conf/conf.toml", "配置文件路径")
	//addr         = flag.String("addr", "tcp://127.0.0.1:1883", "server address")
	//etcdAddrs    = flag.String("etcdAddrs", "127.0.0.1:2379", "etcd address")
)

func init() {
	flag.Parse()
}

func main() {
	if *cmdHelp {
		flag.PrintDefaults()
		return
	}

	// 配置初始化
	if err := InitConfig(*confFilePath); err != nil {
		log.Panic(err)
		return
	}

	u, err := url.Parse(Conf.Server.Addr)
	if err != nil {
		log.Panic(err)
		return
	}

	// @TODO CMD参数覆盖Conf配置

	log.SetColor(true)
	log.SetLevel(Conf.LogLvl())

	if Conf.Auth.Provider == auth.ProviderRpc {
		closer := auth.NewRpcRegister(Conf.App.AccessKey, Conf.App.SecretKey, Conf.Auth.Addrs)
		defer func() {
			if err := closer.Close(); err != nil {
				log.Warnf("rpc auth close error:%v", err)
			}
		}()
	}

	if len(Conf.Server.WsAddr) > 0 {
		service.AddWebsocketHandler("/mqtt", Conf.Server.Addr)
		go service.ListenAndServeWebsocket(Conf.Server.WsAddr)
	}

	server, err := service.NewServer()
	if err != nil {
		log.Panic(err)
		return
	}
	err = server.ListenAndServe(u.Scheme, u.Host)
	if err != nil {
		log.Panic(err)
		return
	}
}
