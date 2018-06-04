package handler

import (
	"context"

	proto "github.com/hb-go/micro-mq/auth/proto/verify"
	"github.com/hb-go/micro-mq/pkg/log"
)

func (a *Auth) Verify(ctx context.Context, req *proto.VerifyReq, resp *proto.VerifyResp) error {
	log.Debugf("rpc: auth verify")
	resp.Verified = false
	if req.Name == req.Pwd {
		resp.Verified = true
	}

	return nil
}
