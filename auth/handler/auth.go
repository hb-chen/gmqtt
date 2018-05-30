package handler

import (
	"context"

	proto "github.com/hb-go/micro-mq/auth/proto"
)

type Auth struct {
}

func (a *Auth) Verify(ctx context.Context, req *proto.Req, resp *proto.Resp) error {
	resp.Verified = false
	if req.Name == req.Pwd {
		resp.Verified = true
	}

	return nil
}
