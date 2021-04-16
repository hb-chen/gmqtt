/*
 * Copyright (c) 2013 IBM Corp.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v10.html
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

// This is originally from
// git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git/samples/simple.go
// I turned it into a test that I can run from `go test`
package benchmark

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"

	"github.com/hb-chen/gmqtt/pkg/log"
)

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var (
	messages    int64         = 10
	publishers  int64         = 100
	subscribers int64         = 1 // =0不消费消息
	size        int           = 64
	topic       []byte        = []byte("topic/1")
	qos         byte          = 1
	order       bool          = true
	nap         int64         = 100
	host        string        = "192.168.1.6"
	port        int           = 1884
	user        string        = "name"
	pass        string        = "pwd"
	version     int           = 4
	pubTimeout  time.Duration = 3000

	subdone, pubdone, rcvdone, sentdone int64

	sReady, pReady         chan struct{}
	rDone, sDone, rTimeOut chan struct{}

	totalSent,
	totalDropped,
	totalSentTime,
	totalRcvd,
	totalRcvdTime,
	sentSince,
	rcvdSince int64

	statMu sync.Mutex
)

func init() {
	log.SetColor(true)
	log.SetLevel(log.INFO)
}

// Usage: go test -run=TestClient$
func TestClient(t *testing.T) {
	rDone = make(chan struct{})

	var wg sync.WaitGroup

	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID("clientId").SetOrderMatters(order)
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	totalDropped = 0

	// 消息消费
	received := int64(0)
	rNow := time.Now()
	rSince := time.Since(rNow).Nanoseconds()
	if subscribers > 0 {
		token = c.Subscribe(string(topic), qos, func(client MQTT.Client, message MQTT.Message) {
			if received == 0 {
				rNow = time.Now()
			}

			received++
			rSince = time.Since(rNow).Nanoseconds()
			log.Debugf("%d MSG: %s\n", received, message.Payload())

			if received >= messages-totalDropped {

				close(rDone)
			}
		})
		token.WaitTimeout(time.Millisecond * 100)
		token.Wait()
		require.NoError(t, token.Error())
	}

	// 消息发布
	now := time.Now()
	since := time.Since(now).Nanoseconds()
	for j := 0; j < 1; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			drop := int64(0)
			payload := make([]byte, size)
			for i := int64(0); i < messages; i++ {
				//wait := time.After(time.Millisecond * 10)
				//<-wait

				msg := fmt.Sprintf("%d", i)
				copy(payload, []byte(msg))
				token = c.Publish(string(topic), qos, false, payload)
				tTimeout := token.WaitTimeout(time.Millisecond * pubTimeout)
				require.NoError(t, token.Error())

				if !tTimeout {
					drop++
					log.Warnf("publish token wait/timeout msg:%v", msg)
				}
			}

			statMu.Lock()
			since = time.Since(now).Nanoseconds()
			totalSent += messages - drop
			totalDropped += drop
			statMu.Unlock()
		}()
	}

	wg.Wait()

	if subscribers > 0 {
		select {
		case <-rDone:
		case <-time.After(time.Millisecond * 1000 * time.Duration(messages)):
			log.Errorf("receiver wait time out")
		}
	}

	time.Sleep(3 * time.Second)

	if subscribers > 0 {
		token = c.Unsubscribe(string(topic))
		if ok := token.WaitTimeout(time.Second * 10); !ok {
			log.Errorf("unsubscribe time out")
		}
		require.NoError(t, token.Error())
	}

	c.Disconnect(250)

	log.Infof("Total sent %d messages dropped %d in %f ms, %f ms/msg, %d msgs/sec", totalSent, totalDropped, float64(since)/float64(time.Millisecond), float64(since)/float64(time.Millisecond)/float64(totalSent), int(float64(totalSent)/(float64(since)/float64(time.Second))))
	log.Infof("Total received %d messages in %f ms, %f ms/msg, %d msgs/sec", received, float64(rSince)/float64(time.Millisecond), float64(rSince)/float64(time.Millisecond)/float64(received), int(float64(received)/(float64(rSince)/float64(time.Second))))
}

// Usage: go test -run=TestClients$
func TestClients(t *testing.T) {

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	totalSent = 0
	totalDropped = 0
	totalRcvd = 0
	totalSentTime = 0
	totalRcvdTime = 0
	sentSince = 0
	rcvdSince = 0

	subdone = 0
	pubdone = 0
	rcvdone = 0
	sentdone = 0

	sReady = make(chan struct{})
	pReady = make(chan struct{})
	rDone = make(chan struct{})
	sDone = make(chan struct{})
	rTimeOut = make(chan struct{})

	if subscribers > 0 {
		for i := int64(1); i < subscribers+1; i++ {
			time.Sleep(time.Millisecond * 20)
			wg1.Add(1)
			go startSubscriber(t, i, &wg1)
		}

		<-sReady
	}

	for i := subscribers + 1; i < publishers+subscribers+1; i++ {
		time.Sleep(time.Millisecond * 20)
		wg2.Add(1)
		go startPublisher(t, i, &wg2)
	}

	wg1.Wait()
	wg2.Wait()

	log.Infof("Total Sent %d messages dropped %d in %f ms, %f ms/msg, %d msgs/sec", totalSent, totalDropped, float64(sentSince)/float64(time.Millisecond), float64(sentSince)/float64(time.Millisecond)/float64(totalSent), int(float64(totalSent)/(float64(sentSince)/float64(time.Second))))
	log.Infof("Total Received %d messages in %f ms, %f ms/msg, %d msgs/sec", totalRcvd, float64(sentSince)/float64(time.Millisecond), float64(sentSince)/float64(time.Millisecond)/float64(totalRcvd), int(float64(totalRcvd)/(float64(sentSince)/float64(time.Second))))
}

func startSubscriber(t testing.TB, cid int64, wg *sync.WaitGroup) {
	defer wg.Done()

	clientId := fmt.Sprintf("clientId-%d", cid)
	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID(clientId).SetOrderMatters(order)
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	ok := token.WaitTimeout(time.Second * 10)
	if !ok {
		log.Errorf("subscriber connect time out")
		require.NoError(t, errors.New("subscriber connect time out"))
		return
	}
	require.NoError(t, token.Error())

	now := time.Now()
	since := time.Since(now).Nanoseconds()
	cnt := messages * publishers
	received := int64(0)
	done := int64(0)

	filters := map[string]byte{string(topic): qos}
	token = c.SubscribeMultiple(filters, func(client MQTT.Client, message MQTT.Message) {
		if received == 0 {
			now = time.Now()
		}

		received++
		since = time.Since(now).Nanoseconds()

		log.Debugf("(cid:%d) messages received:%d, msg:%v", cid, received, string(message.Payload()))
		if received%int64(float64(cnt)*0.1) == 0 {
			log.Infof("(cid:%d) messages received:%g%%", cid, float64(received)/float64(cnt)*100.0)
		}

		if done == 0 && received >= cnt-totalDropped {
			doit := atomic.CompareAndSwapInt64(&done, 0, 1)
			if doit {
				rcvd := atomic.AddInt64(&rcvdone, 1)
				if rcvd == subscribers {
					close(rDone)
				}
			}
		}
	})
	token.WaitTimeout(time.Millisecond * 100)
	token.Wait()
	require.NoError(t, token.Error())

	log.Infof("(cid:%d) subscriber ready", cid)

	subs := atomic.AddInt64(&subdone, 1)
	if subs == subscribers {
		log.Infof("subscribers sub done")
		close(sReady)
	}

	<-sDone

	allDone := false
	if done == 0 && received >= cnt-totalDropped {
		doit := atomic.CompareAndSwapInt64(&done, 0, 1)
		if doit {
			rcvd := atomic.AddInt64(&rcvdone, 1)
			if rcvd == subscribers {
				close(rDone)
				allDone = true
			}
		} else {
			if rcvdone == subscribers {
				allDone = true
			}
		}
	}

	if !allDone {
		select {
		case <-rDone:
		case <-time.After(time.Millisecond * time.Duration(nap*publishers)):
			close(rTimeOut)
			log.Warnf("(cid:%d) Timed out waiting for messages to be received.", cid)
		}
	}

	token = c.Unsubscribe(string(topic))
	if ok := token.WaitTimeout(time.Second * 5); !ok {
		log.Errorf("subscriber unsub time out")
	}
	require.NoError(t, token.Error())

	time.Sleep(3 * time.Second)
	c.Disconnect(250)

	statMu.Lock()
	totalRcvd += int64(received)
	totalRcvdTime += int64(since)
	if since > rcvdSince {
		rcvdSince = since
	}
	statMu.Unlock()

	log.Infof("(cid:%d) Received %d messages in %f ms, %f ms/msg, %d msgs/sec", cid, received, float64(since)/float64(time.Millisecond), float64(since)/float64(time.Millisecond)/float64(cnt), int(float64(received)/(float64(since)/float64(time.Second))))
}

func startPublisher(t testing.TB, cid int64, wg *sync.WaitGroup) {
	defer wg.Done()

	clientId := fmt.Sprintf("clientId-%d", cid)
	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID(clientId).SetOrderMatters(order).SetMessageChannelDepth(1000)
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	ok := token.WaitTimeout(time.Millisecond * 1000)
	if !ok {
		log.Errorf("publisher connect time out")
		require.NoError(t, errors.New("publisher connect time out"))
		return
	}
	require.NoError(t, token.Error())

	pubs := atomic.AddInt64(&pubdone, 1)
	log.Infof("(cid:%d) publisher ready, total:%d (%d)", cid, pubs, cid-pubs)
	if pubs == int64(publishers) {
		log.Infof("publishers pub done")
		close(pReady)
	} else {
		<-pReady
	}

	now := time.Now()
	since := time.Since(now).Nanoseconds()
	cnt := messages
	sent := 0
	dropped := 0
	payload := make([]byte, size)
	for i := int64(0); i < cnt; i++ {
		// 随机延时发送
		if false {
			r := int64(100)
			if false {
				r += rand.Int63n(100)
			}
			wait := time.After(time.Millisecond * time.Duration(r))
			<-wait
		}

		msg := fmt.Sprintf("cid:%d, msg:%d", cid, i)
		copy(payload, []byte(msg))
		token = c.Publish(string(topic), qos, false, payload)
		tTimeout := token.WaitTimeout(time.Millisecond * pubTimeout)
		require.NoError(t, token.Error())

		if !tTimeout {
			dropped++
			log.Warnf("(cid:%d) publish wait timeout", cid)
		} else {
			sent++
		}

		since = time.Since(now).Nanoseconds()
	}

	statMu.Lock()
	totalDropped += int64(dropped)
	statMu.Unlock()

	sends := atomic.AddInt64(&sentdone, 1)
	log.Infof("(cid:%d) publish done, total:%d", cid, sends)
	if sends == int64(publishers) {
		log.Infof("publishers sent done")
		close(sDone)
	}

	if subscribers > 0 {
		select {
		case <-rDone:
		case <-rTimeOut:
		}
	}

	time.Sleep(3 * time.Second)
	c.Disconnect(250)

	statMu.Lock()
	totalSent += int64(sent)
	totalSentTime += int64(since)
	if since > sentSince {
		sentSince = since
	}
	statMu.Unlock()

	log.Infof("(cid:%d) Sent %d messages dropped %d in %f ms, %f ms/msg, %d msgs/sec", cid, sent, dropped, float64(since)/float64(time.Millisecond), float64(since)/float64(time.Millisecond)/float64(cnt), int(float64(sent)/(float64(since)/float64(time.Second))))
}
