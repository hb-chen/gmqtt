// +build rabbitmq
package rabbitmq

import (
	"context"

	"github.com/hb-chen/micro-mq/broker"
)

type durableQueueKey struct{}
type headersKey struct{}
type exchangeKey struct{}

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, durableQueueKey{}, true)
	}
}

// Headers adds headers used by the headers exchange
func Headers(h map[string]interface{}) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, headersKey{}, h)
	}
}

// Exchange is an option to set the Exchange
func Exchange(e string) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, exchangeKey{}, e)
	}
}
