package client

import (
	"github.com/smallnest/rpcx/server"

	"github.com/hb-go/micro-mq/api/client/auth"
	"github.com/hb-go/micro-mq/api/client/register"
)

func Register(s *server.Server) {
	auth.Register(s)
	register.Register(s)
}
