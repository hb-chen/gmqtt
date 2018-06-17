package client

import (
	"context"

	pbRgt "github.com/hb-go/micro-mq/api/client/proto/register"
	pbUnrgt "github.com/hb-go/micro-mq/api/client/proto/unregister"
)

func (*Client) Register(ctx context.Context, req *pbRgt.RegisterReq, resp *pbRgt.RegisterResp) error {
	if req.ClientId != "" {
		resp.ClientId = "client_id"
	} else {
		resp.ClientId = req.ClientId
	}

	resp.Name = "name"
	resp.Pwd = "pwd"

	return nil
}

func (*Client) Unregister(ctx context.Context, req *pbUnrgt.UnregisterReq, resp *pbUnrgt.UnregisterResp) error {
	if req.ClientId != "" {
		resp.ClientId = "client_id"
	} else {
		resp.ClientId = req.ClientId
	}

	return nil
}
