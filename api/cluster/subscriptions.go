package cluster

import (
	"context"

	pbClients "github.com/hb-go/micro-mq/api/cluster/proto/clients"
	pb "github.com/hb-go/micro-mq/api/cluster/proto/subscriptions"
	pbTopics "github.com/hb-go/micro-mq/api/cluster/proto/topics"
)

func (*cluster) Subscriptions(ctx context.Context, req *pb.SubscriptionsReq, resp *pb.SubscriptionsResp) error {
	// @TODO mock
	for i := int64(0); i < req.Size; i++ {
		c := &pbClients.Client{
			Id: "id",
		}

		t := &pbTopics.Topic{
			Topic: "topic",
			Qos:   0,
		}

		s := &pb.Subscriptition{
			Client: c,
			Topic:  t,
		}

		resp.Subscriptions = append(resp.Subscriptions, s)
	}
	return nil
}
