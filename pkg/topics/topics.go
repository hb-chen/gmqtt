// Copyright (c) 2014 The SurgeMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package topics deals with MQTT topic names, topic filters and subscriptions.
// - "Topic name" is a / separated string that could contain #, * and $
// - / in topic name separates the string into "topic levels"
// - # is a multi-level wildcard, and it must be the last character in the
//   topic name. It represents the parent and all children levels.
// - + is a single level wildwcard. It must be the only character in the
//   topic level. It represents all names in the current level.
// - $ is a special character that says the topic is a system level topic
package topics

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/surgemq/message"

	"github.com/hb-chen/gmqtt/pkg/broker"
	. "github.com/hb-chen/gmqtt/pkg/conf"
	"github.com/hb-chen/gmqtt/pkg/log"
)

const (
	// MWC is the multi-level wildcard
	MWC = "#"

	// SWC is the single level wildcard
	SWC = "+"

	// SEP is the topic level separator
	SEP = "/"

	// SYS is the starting character of the system level topics
	SYS = "$"

	// Both wildcards
	_WC = "#+"

	// MQTT <=> MQ
	MQHeaderMQTTQos   = "MQTT_Qos"
	MQHeaderMQTTTopic = "MQTT_Topic"
)

var (
	// ErrAuthFailure is returned when the user/pass supplied are invalid
	ErrAuthFailure = errors.New("auth: Authentication failure")

	// ErrAuthProviderNotFound is returned when the requested provider does not exist.
	// It probably hasn't been registered yet.
	ErrAuthProviderNotFound = errors.New("auth: Authentication provider not found")

	providers = make(map[string]TopicsProvider)
)

// TopicsProvider
type TopicsProvider interface {
	Subscribe(topic []byte, qos byte, subscriber interface{}) (byte, error)
	Unsubscribe(topic []byte, subscriber interface{}) (int, error)
	Subscribers(topic []byte, qos byte, subs *[]interface{}, qoss *[]byte) error
	Retain(msg *message.PublishMessage) error
	Retained(topic []byte, msgs *[]*message.PublishMessage) error
	Close() error
}

func Register(name string, provider TopicsProvider) {
	if provider == nil {
		panic("topics: Register provide is nil")
	}

	if _, dup := providers[name]; dup {
		panic("topics: Register called twice for provider " + name)
	}

	providers[name] = provider
}

func Unregister(name string) {
	delete(providers, name)
}

type Manager struct {
	p TopicsProvider

	broker broker.Broker
	subers sync.Map

	subHandler broker.Handler

	subs []interface{}
	qoss []byte
}

func NewManager(providerName string, b broker.Broker, h broker.Handler) (*Manager, error) {
	p, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("topics: unknown provider %q", providerName)
	}

	return &Manager{
		p:          p,
		broker:     b,
		subHandler: h,
	}, nil
}

func (this *Manager) Subscribe(topic []byte, qos byte, subscriber interface{}) (byte, error) {
	// @TODO 订阅鉴权

	// broker订阅
	// 检查topic是否已订阅
	// broker topic匹配规则转换
	if this.broker != nil && this.subHandler != nil {
		brTopic := TopicToBrokerTopic(topic)
		log.Debugf("broker topic:%v subscribe", brTopic)
		_, ok := this.subers.Load(brTopic)
		if !ok {
			// broker只订阅一级话题，通过header传递MQTT特有属性QoS、Topic
			// @TODO Broker队列与Gateway实例ID绑定，即Kafka的Consumer Group
			suber, err := this.broker.Subscribe(brTopic, this.subHandler, broker.Queue(Conf.Server.Id))
			if err != nil {
				log.Errorf("broker topic:%v subscribe error:%v", brTopic, err)
				return message.QosFailure, err
			} else {
				log.Infof("broker topic:%v subscribed", brTopic)
				this.subers.Store(brTopic, suber)
			}
		}
	}

	log.Debugf("topic:%v subscribe", string(topic))
	qos, err := this.p.Subscribe(topic, qos, subscriber)
	if err != nil {
		log.Errorf("topic:%v subscribe error:%v", string(topic), err)
		// @TODO 节点内订阅失败，broker是否需要取消订阅
	}

	return qos, err
}

func (this *Manager) Unsubscribe(topic []byte, subscriber interface{}) error {
	nodesNum, err := this.p.Unsubscribe(topic, subscriber)
	if err != nil {
		// @TODO 排除"No topic found"错误
		return err
	}

	// 节点内取消订阅后刷新broker订阅
	// topic下所有snodes数为0时取消订阅
	if this.broker != nil && this.subHandler != nil {
		if nodesNum <= 0 {
			brTopic := TopicToBrokerTopic(topic)
			v, ok := this.subers.Load(brTopic)
			if ok {
				suber, ok := v.(broker.Subscriber)
				if ok {
					if err = suber.Unsubscribe(); err != nil {
						return err
					}
				}

				this.subers.Delete(brTopic)
			}
		}
	}

	return err
}

func (this *Manager) Subscribers(topic []byte, qos byte, subs *[]interface{}, qoss *[]byte) error {
	return this.p.Subscribers(topic, qos, subs, qoss)
}

func (this *Manager) Retain(msg *message.PublishMessage) error {
	return this.p.Retain(msg)
}

func (this *Manager) Retained(topic []byte, msgs *[]*message.PublishMessage) error {
	return this.p.Retained(topic, msgs)
}

func (this *Manager) Close() error {
	return this.p.Close()
}

func TopicToBrokerTopic(topic []byte) string {
	// @TODO topic对应关系，n:1时Unsubscribe()需要重新设计
	str := string(topic)
	subs := strings.Split(str, "/")

	//log.Errorf("topic subs:%v", subs)

	if len(subs) > 0 {
		return strings.Trim(subs[0], "/")
	} else {
		// @TODO 默认topic或增加ERR
		return "topic"
	}
}

func BrokerTopicToTopic(topic string) []byte {
	topic = strings.Replace(topic, ".", "/", -1)

	return []byte(topic)
}
