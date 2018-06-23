package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/persist/file-adapter"
	"github.com/casbin/casbin/persist"
	"github.com/smallnest/rpcx/protocol"

	"github.com/hb-go/micro-mq/pkg/util/crypt"
)

type Auth struct {
	cse *casbin.SyncedEnforcer
}

func NewAuth(adapter persist.Adapter) Auth {
	var e *casbin.SyncedEnforcer
	if adapter != nil {
		e = casbin.NewSyncedEnforcer("conf/casbin/rbac_with_deny_model.conf", adapter, false)
	} else {
		e = casbin.NewSyncedEnforcer("conf/casbin/rbac_with_deny_model.conf", "conf/casbin/rbac_with_deny_policy.csv", false)
	}

	a := Auth{
		cse: e,
	}

	return a
}

func (this Auth) Init() error {
	err := this.cse.LoadPolicy()
	if err != nil {
		return err
	}

	filter := fileadapter.Filter{
		P: []string{"a", "b"},
		G: []string{"a", "b"},
	}
	this.cse.LoadFilteredPolicy(filter)

	this.cse.StartAutoLoadPolicy(time.Second * 60)

	return nil
}

func Token(ak, sk, path string) string {
	return "1.0:" + ak + ":" + crypt.MD5([]byte(path+sk))
}

func (this Auth) Verify(ctx context.Context, req *protocol.Message, token string) error {
	sp := strings.Split(token, ":")
	if len(sp) != 3 {
		return errors.New("invalid token")
	}

	ak := sp[1]
	sk := "sk"
	if token != Token(ak, sk, req.ServicePath) {
		return errors.New("invalid token")
	}

	//sub := "alice" // the user that wants to access a resource.
	//obj := "data1" // the resource that is going to be accessed.
	//act := "read"  // the operation that the user performs on the resource.
	if this.cse.Enforce(ak, req.ServicePath, req.ServiceMethod) != true {
		// deny the request, show an error
		return errors.New("deny the request")
	}

	return nil

	// AK/SK查询

	// token验证

	// service path/method授权策略

	// message.Payload验签
}
