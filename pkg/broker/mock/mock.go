package mock

import (
	"github.com/hb-chen/gmqtt/pkg/broker"
)

func NewBroker(opts ...broker.Option) broker.Broker {
	return broker.NewBroker(opts...)
}
