package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/persist"
	"github.com/casbin/casbin/persist/file-adapter"
	"github.com/smallnest/rpcx/protocol"

	"github.com/hb-chen/gmqtt/pkg/log"
	"github.com/hb-chen/gmqtt/pkg/util/crypt"
)

type Auth struct {
	cse *casbin.SyncedEnforcer
}

func Token(ak, sk, path string) string {
	return "1.0:" + ak + ":" + crypt.MD5([]byte(path+sk))
}

func NewAuth(adapter persist.Adapter, w persist.Watcher) *Auth {
	e := casbin.NewSyncedEnforcer("conf/casbin/rbac_with_deny_model.conf", adapter, false)
	a := &Auth{
		cse: e,
	}

	if w != nil {
		e.SetWatcher(w)
		w.SetUpdateCallback(a.casbinUpdateCallback)
	}

	return a
}

func (this *Auth) Init() error {
	err := this.cse.LoadPolicy()
	if err != nil {
		return err
	}

	// @TODO
	if false {
		filter := fileadapter.Filter{
			P: []string{"a", "b"},
			G: []string{"a", "b"},
		}
		this.cse.LoadFilteredPolicy(filter)
	}

	this.cse.StartAutoLoadPolicy(time.Second * 60)

	return nil
}

func (this *Auth) Verify(ctx context.Context, req *protocol.Message, token string) error {
	// AK/SK查询
	// token验证
	// service path/method授权策略
	// message.Payload验签

	// {token版本/加密方式}:{ak}:{token加密}
	sp := strings.Split(token, ":")
	if len(sp) != 3 {
		return errors.New("invalid token split len != 3")
	}

	// AK/SK查询
	// @TODO AccessKey存储引擎
	ak := AccessKey{
		AccessKeyId:     sp[1],
		AccessKeySecret: "sk",
		Roles:           []string{"client"},
	}

	// token验证
	if token != Token(ak.AccessKeyId, ak.AccessKeySecret, req.ServicePath) {
		return errors.New("invalid token")
	}

	// service path/method授权策略
	for _, v := range ak.Roles {
		if this.cse.Enforce(v, req.ServicePath, req.ServiceMethod) == true {
			return nil
		}
	}

	// deny the request, show an error
	return errors.New("deny the request")
}

func (this *Auth) casbinUpdateCallback(rev string) {
	log.Infof("new revision detected:", rev)
	this.cse.LoadPolicy()
}
