package auth

import (
	"context"
	"io"

	"github.com/smallnest/rpcx/client"

	"github.com/hb-go/micro-mq/pkg/util/conv"
	"github.com/hb-go/micro-mq/pkg/log"
	auth "github.com/hb-go/micro-mq/auth/proto"
	verify "github.com/hb-go/micro-mq/auth/proto/verify"
)

const (
	ProviderRpc = "rpc"
)

type RpcAuthenticator struct {
	xClient  client.XClient
	EtcdAddr []string
}

func NewRpcRegister(addr []string) io.Closer {
	rpcAuth := &RpcAuthenticator{EtcdAddr: addr}
	rpcAuth.Init()
	Register(ProviderRpc, rpcAuth)

	return rpcAuth
}

func (a *RpcAuthenticator) Init() {
	d := client.NewEtcdDiscovery(conv.ProtoEnumsToRpcxBasePath(auth.BASE_PATH_name), auth.SRV_Auth.String(), a.EtcdAddr, nil)
	xc := client.NewXClient(auth.SRV_Auth.String(), client.Failover, client.RoundRobin, d, client.DefaultOption)
	a.xClient = xc
}

func (a *RpcAuthenticator) Authenticate(id string, cred interface{}) error {

	if pwd, ok := cred.(string); !ok {
		return ErrAuthCredType
	} else {
		req := &verify.VerifyReq{
			Name: id,
			Pwd:  pwd,
		}
		resp := &verify.VerifyResp{}

		if err := a.xClient.Call(context.Background(), auth.METHOD_Verify.String(), req, resp); err != nil {
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
