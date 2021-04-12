package cluster

import (
	"context"

	pb "github.com/hb-chen/micro-mq/api/cluster/proto/sessions"
)

func (*cluster) Sessions(ctx context.Context, req *pb.SessionsReq, resp *pb.SessionsResp) error {
	// @TODO mock
	for i := int64(0); i < req.Size; i++ {
		s := &pb.Session{
			Id: "id",
		}

		resp.Sessions = append(resp.Sessions, s)
	}
	return nil
}
