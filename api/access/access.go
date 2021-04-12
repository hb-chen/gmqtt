package access

import (
	"github.com/smallnest/rpcx/server"
	//pbAccess "github.com/hb-chen/micro-mq/api/access/proto"
)

type Access struct {
}

func Register(s *server.Server) {
	//s.RegisterName(pbAccess.SRV_access.String(), new(Access), "")
}
