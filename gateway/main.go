package main

import (
	"flag"
	"net/url"

	"github.com/hb-go/micro-mq/pkg/log"
	"github.com/hb-go/micro-mq/gateway/auth"
	"github.com/hb-go/micro-mq/gateway/service"
)

var (
	cmdHelp  = flag.Bool("h", false, "帮助")
	addr     = flag.String("addr", "tcp://127.0.0.1:1883", "server address")
	etcdAddr = flag.String("etcdAddr", "localhost:2379", "etcd address")
)

func init() {
	flag.Parse()
}

func main() {
	if *cmdHelp {
		flag.PrintDefaults()
		return
	}

	u, err := url.Parse(*addr)
	if err != nil {
		log.Panic(err)
		return
	}

	log.SetColor(true)
	log.SetLevel(log.DEBUG)

	closer := auth.NewRpcRegister(*etcdAddr)
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warnf("rpc auth close error:%v", err)
		}
	}()

	service.AddWebsocketHandler("/mqtt", *addr)
	go service.ListenAndServeWebsocket(":8080")

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
