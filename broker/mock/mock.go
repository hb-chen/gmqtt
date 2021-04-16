package mock

import (
	"github.com/hb-chen/gmqtt/broker"
)

func NewBroker(opts ...broker.Option) broker.Broker {
	return broker.NewBroker(opts...)
}
