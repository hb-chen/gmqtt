package auth

import (
	"context"
	"errors"

	"github.com/smallnest/rpcx/protocol"
	"github.com/hb-go/micro-mq/pkg/util/crypt"
	"strings"
)

func Token(ak, sk, path string) string {
	return "1.0:" + ak + ":" + crypt.MD5([]byte(path+sk))
}

func Verify(ctx context.Context, req *protocol.Message, token string) error {
	sp := strings.Split(token, ":")
	if len(sp) != 3 {
		return errors.New("invalid token")
	}

	ak := sp[1]
	sk := "sk"
	if token != Token(ak, sk, req.ServicePath) {
		return errors.New("invalid token")
	}

	return nil

	// AK/SK查询

	// token验证

	// service path/method授权策略

	// message.Payload验签
}
