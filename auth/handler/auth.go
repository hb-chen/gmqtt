package handler

import (
	"github.com/smallnest/rpcx/server"

	"github.com/hb-go/micro-mq/auth/proto"
)

type Auth struct {
}

func Register(s *server.Server) {
	s.RegisterName(proto.SRV_Auth.String(), new(Auth), "")
}
