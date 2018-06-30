package register

import (
	"context"

	"github.com/smallnest/rpcx/server"

	client "github.com/hb-go/micro-mq/api/client/proto"
	pb "github.com/hb-go/micro-mq/api/client/register/proto"
)

func Register(s *server.Server) {
	s.RegisterName(client.SRV_client_register.String(), new(register), "")
}

type register struct {
}

func (*register) Register(ctx context.Context, req *pb.RegisterReq, resp *pb.RegisterResp) error {
	if req.ClientId != "" {
		resp.ClientId = "client_id"
	} else {
		resp.ClientId = req.ClientId
	}

	resp.Name = "name"
	resp.Pwd = "pwd"

	return nil
}

func (*register) Unregister(ctx context.Context, req *pb.UnregisterReq, resp *pb.UnregisterResp) error {
	if req.ClientId != "" {
		resp.ClientId = "client_id"
	} else {
		resp.ClientId = req.ClientId
	}

	return nil
}
