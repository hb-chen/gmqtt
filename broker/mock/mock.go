package mock

import (
	"github.com/hb-chen/micro-mq/broker"
)

func NewBroker(opts ...broker.Option) broker.Broker {
	return broker.NewBroker(opts...)
}
