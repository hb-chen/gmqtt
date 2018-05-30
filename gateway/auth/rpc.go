package auth

import (
	"context"

	"github.com/smallnest/rpcx/client"

	"github.com/hb-go/micro-mq/pkg/util/conv"
	"github.com/hb-go/micro-mq/pkg/log"
	auth "github.com/hb-go/micro-mq/auth/proto"
)

var (
	ProviderNameRpc = "rpc"
)

type RpcAuthenticator struct {
	EtcdAddr *string
}

func RegistNewRpc(addr string) {
	Register(ProviderNameRpc, &RpcAuthenticator{EtcdAddr: &addr})
}

func (a *RpcAuthenticator) Authenticate(id string, cred interface{}) error {
	d := client.NewEtcdDiscovery(conv.ProtoEnumsToRpcxBasePath(auth.BASE_PATH_name), auth.SRV_Auth.String(), []string{*a.EtcdAddr}, nil)
	xclient := client.NewXClient(auth.SRV_Auth.String(), client.Failover, client.RoundRobin, d, client.DefaultOption)
	defer xclient.Close()

	if pwd, ok := cred.(string); !ok {
		return ErrAuthCredType
	} else {
		req := &auth.Req{
			Name: id,
			Pwd:  pwd,
		}
		resp := &auth.Resp{}
		err := xclient.Call(context.Background(), auth.METHOD_Verify.String(), req, resp)
		if err != nil {
			log.Panic(err)
		}

		return nil
	}

}
