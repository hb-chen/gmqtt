package client

import (
	"github.com/smallnest/rpcx/server"

	pbClient "github.com/hb-go/micro-mq/api/client/proto"
)

type Client struct {
}

func Register(s *server.Server) {
	s.RegisterName(pbClient.SRV_client.String(), new(Client), "")
}
