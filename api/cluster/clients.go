package cluster

import (
	"context"

	pb "github.com/hb-chen/micro-mq/api/cluster/proto/clients"
)

func (*cluster) Clients(ctx context.Context, req *pb.ClientsReq, resp *pb.ClientsResp) error {
	// @TODO mock
	for i := int64(0); i < req.Size; i++ {
		c := &pb.Client{
			Id: "id",
		}

		resp.Clients = append(resp.Clients, c)
	}
	return nil
}
