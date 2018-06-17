package auth

import (
	"context"
	"io"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"

	"github.com/hb-go/micro-mq/api/auth"
	"github.com/hb-go/micro-mq/pkg/util/conv"
	"github.com/hb-go/micro-mq/pkg/log"
	pbApi "github.com/hb-go/micro-mq/api/proto"
	pbClient "github.com/hb-go/micro-mq/api/client/proto"
	pbAuth "github.com/hb-go/micro-mq/api/client/proto/auth"
)

const (
	ProviderRpc = "rpc"
)

type RpcAuthenticator struct {
	AccessKey string
	SecretKey string
	xClient   client.XClient
	etcdAddr  []string
}

func NewRpcRegister(ak, sk string, addr []string) io.Closer {
	rpcAuth := &RpcAuthenticator{
		AccessKey: ak,
		SecretKey: sk,
		etcdAddr:  addr,
	}
	rpcAuth.Init()
	Register(ProviderRpc, rpcAuth)

	return rpcAuth
}

func (a *RpcAuthenticator) Init() {
	d := client.NewEtcdDiscovery(conv.ProtoEnumsToRpcxBasePath(pbApi.BASE_PATH_name), pbClient.SRV_client.String(), a.etcdAddr, nil)
	xc := client.NewXClient(pbClient.SRV_client.String(), client.Failover, client.RoundRobin, d, client.DefaultOption)
	xc.Auth(auth.Token(a.AccessKey, a.SecretKey, pbClient.SRV_client.String()))
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
		if err := a.xClient.Call(ctx, pbClient.METHOD_Auth.String(), req, resp); err != nil {
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
