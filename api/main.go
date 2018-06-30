package main

import (
	"flag"
	"time"

	//"github.com/casbin/redis-adapter"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/casbin/casbin/persist/file-adapter"
	"github.com/casbin/etcd-watcher"
	"github.com/casbin/redis-adapter"

	"github.com/hb-go/micro-mq/pkg/log"
	"github.com/hb-go/micro-mq/pkg/util/conv"
	"github.com/hb-go/micro-mq/api/auth"
	"github.com/hb-go/micro-mq/api/proto"
	"github.com/hb-go/micro-mq/api/access"
	"github.com/hb-go/micro-mq/api/client"
	"github.com/hb-go/micro-mq/api/cluster"
)

var (
	cmdHelp  = flag.Bool("h", false, "帮助")
	addr     = flag.String("addr", "127.0.0.1:8972", "server address")
	etcdAddr = flag.String("etcdAddr", "127.0.0.1:2379", "etcd address")
)

func init() {
	flag.Parse()
}

func main() {
	if *cmdHelp {
		flag.PrintDefaults()
		return
	}

	log.SetColor(true)
	log.SetLevel(log.DEBUG)

	s := server.NewServer()
	addRegistryPlugin(s)

	access.Register(s)
	client.Register(s)
	cluster.Register(s)

	// Auth
	var a *auth.Auth
	if false {
		// @TODO redis adapter, etcd watcher管理
		adapter := redisadapter.NewAdapter("tcp", "127.0.0.1:6379")
		w := etcdwatcher.NewWatcher("http://127.0.0.1:2379")
		a = auth.NewAuth(adapter, w)
	} else {
		adapter := fileadapter.NewAdapter("conf/casbin/rbac_with_deny_policy.csv")
		a = auth.NewAuth(adapter, nil)
	}
	if err := a.Init(); err != nil {
		log.Fatalf("api client serve start error:", err)
	}
	s.AuthFunc = a.Verify

	err := s.Serve("tcp", *addr)
	if err != nil {
		log.Fatalf("api client serve exit error:%v", err)
	}
}

func addRegistryPlugin(s *server.Server) {
	r := &serverplugin.EtcdRegisterPlugin{
		ServiceAddress: "tcp@" + *addr,
		EtcdServers:    []string{*etcdAddr},
		BasePath:       conv.ProtoEnumsToRpcxBasePath(proto.BASE_PATH_name),
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.Plugins.Add(r)
}
