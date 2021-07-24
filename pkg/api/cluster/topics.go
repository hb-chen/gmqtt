package cluster

import (
	"context"

	pb "github.com/hb-chen/gmqtt/pkg/api/cluster/proto/topics"
)

func (*cluster) Topics(ctx context.Context, req *pb.TopicsReq, resp *pb.TopicsResp) error {
	// @TODO mock
	for i := int64(0); i < req.Size; i++ {
		t := &pb.Topic{
			Topic: "topic",
			Qos:   0,
		}

		resp.Topics = append(resp.Topics, t)
	}
	return nil
}
