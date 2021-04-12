package auth

import (
	"context"
	"io"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"

	"github.com/hb-chen/micro-mq/api/auth"
	pbAuth "github.com/hb-chen/micro-mq/api/client/auth/proto"
	pbClient "github.com/hb-chen/micro-mq/api/client/proto"
	pbApi "github.com/hb-chen/micro-mq/api/proto"
	"github.com/hb-chen/micro-mq/pkg/log"
	"github.com/hb-chen/micro-mq/pkg/util/conv"
)

const (
	ProviderRpc = "rpc"
)

type RpcAuthenticator struct {
	AccessKeyId     string
	AccessKeySecret string
	xClient         client.XClient
	consulAddr      []string
}

func NewRpcRegister(ak, sk string, addr []string) io.Closer {
	rpcAuth := &RpcAuthenticator{
		AccessKeyId:     ak,
		AccessKeySecret: sk,
		consulAddr:      addr,
	}
	rpcAuth.init()
	Register(ProviderRpc, rpcAuth)

	return rpcAuth
}

func (a *RpcAuthenticator) init() {
	d, err := client.NewConsulDiscovery(
		conv.ProtoEnumsToRpcxBasePath(pbApi.BASE_PATH_name),
		pbClient.SRV_client_auth.String(),
		a.consulAddr,
		nil,
	)
	if err != nil {
		panic(err)
	}
	xc := client.NewXClient(pbClient.SRV_client_auth.String(), client.Failover, client.RoundRobin, d, client.DefaultOption)
	xc.Auth(auth.Token(a.AccessKeyId, a.AccessKeySecret, pbClient.SRV_client_auth.String()))
	a.xClient = xc
}

func (a *RpcAuthenticator) Authenticate(id string, cred interface{}) error {
	if pwd, ok := cred.(string); !ok {
		return ErrAuthCredType
	} else {
		req := &pbAuth.AuthReq{
			Name: id,
			Pwd:  pwd,
		}
		resp := &pbAuth.AuthResp{}
		ctx := context.WithValue(context.Background(), share.ReqMetaDataKey, make(map[string]string))
		if err := a.xClient.Call(ctx, pbAuth.METHOD_Auth.String(), req, resp); err != nil {
			log.Panic(err)
		}

		if !resp.Verified {
			return ErrAuthFailure
		} else {
			return nil
		}
	}
}

func (a *RpcAuthenticator) Close() error {
	return a.xClient.Close()
}
