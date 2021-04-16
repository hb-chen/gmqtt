package auth

import (
	"context"

	"github.com/smallnest/rpcx/server"

	pb "github.com/hb-chen/gmqtt/api/client/auth/proto"
	client "github.com/hb-chen/gmqtt/api/client/proto"
	"github.com/hb-chen/gmqtt/pkg/log"
)

func Register(s *server.Server) {
	s.RegisterName(client.SRV_client_auth.String(), new(auth), "")
}

type auth struct {
}

func (*auth) Auth(ctx context.Context, req *pb.AuthReq, resp *pb.AuthResp) error {
	log.Debugf("rpc: auth verify")
	resp.Verified = false
	if req.Name == req.Pwd {
		resp.Verified = true
	}

	return nil
}

func (*auth) SubAuth(ctx context.Context, req *pb.TopicReq, resp *pb.TopicResp) error {
	resp.Allow = true

	return nil
}

func (*auth) PubAuth(ctx context.Context, req *pb.TopicReq, resp *pb.TopicResp) error {
	resp.Allow = true

	return nil
}
