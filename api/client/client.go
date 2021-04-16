package client

import (
	"github.com/smallnest/rpcx/server"

	"github.com/hb-chen/gmqtt/api/client/auth"
	"github.com/hb-chen/gmqtt/api/client/register"
)

func Register(s *server.Server) {
	auth.Register(s)
	register.Register(s)
}
