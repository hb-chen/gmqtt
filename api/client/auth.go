package client

import (
	"context"

	"github.com/hb-go/micro-mq/pkg/log"

	pbAuth "github.com/hb-go/micro-mq/api/client/proto/auth"
)

func (*Client) Auth(ctx context.Context, req *pbAuth.AuthReq, resp *pbAuth.AuthResp) error {
	log.Debugf("rpc: auth verify")
	resp.Verified = false
	if req.Name == req.Pwd {
		resp.Verified = true
	}

	return nil
}
