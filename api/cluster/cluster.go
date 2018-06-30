package cluster

import (
	"github.com/smallnest/rpcx/server"

	pb "github.com/hb-go/micro-mq/api/cluster/proto"
)

type cluster struct {
}

func Register(s *server.Server) {
	s.RegisterName(pb.SRV_cluster.String(), new(cluster), "")
}

