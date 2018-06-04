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
	"fmt"
	"sync"
	"testing"
	"time"
	"math/rand"
	"sync/atomic"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/require"

	"github.com/hb-go/micro-mq/pkg/log"
)

var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

var (
	messages    int    = 10
	publishers  int    = 2
	subscribers int    = 2
	size        int    = 20
	topic       []byte = []byte("topic/1")
	qos         byte   = 0
	nap         int    = 10
	host        string = "127.0.0.1"
	port        int    = 1883
	user        string = "surgemq"
	pass        string = "surgemq"
	version     int    = 4

	subdone, rcvdone, sentdone int64

	done, done2 chan struct{}

	totalSent,
	totalSentTime,
	totalRcvd,
	totalRcvdTime,
	sentSince,
	rcvdSince int64

	statMu sync.Mutex
)

func TestClient(t *testing.T) {

	var wg sync.WaitGroup

	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID("clientId")
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	filters := map[string]byte{string(topic): qos}
	token = c.SubscribeMultiple(filters, nil)
	token.WaitTimeout(time.Millisecond * 100)
	token.Wait()
	require.NoError(t, token.Error())

	for i := 0; i < 100; i++ {
		text := fmt.Sprintf("this is msg #%d!", i)
		token = c.Publish(string(topic), qos, false, []byte(text))
		token.WaitTimeout(time.Millisecond * 100)
		token.Wait()
		require.NoError(t, token.Error())
	}

	time.Sleep(3 * time.Second)

	token = c.Unsubscribe(string(topic))
	token.Wait()
	token.WaitTimeout(time.Millisecond * 100)
	require.NoError(t, token.Error())

	c.Disconnect(250)

	wg.Wait()
}

func TestClients(t *testing.T) {
	log.SetColor(true)
	log.SetLevel(log.INFO)

	var wg sync.WaitGroup

	totalSent = 0
	totalRcvd = 0
	totalSentTime = 0
	totalRcvdTime = 0
	sentSince = 0
	rcvdSince = 0

	subdone = 0
	rcvdone = 0

	done = make(chan struct{})
	done2 = make(chan struct{})

	for i := 1; i < subscribers+1; i++ {
		time.Sleep(time.Millisecond * 20)
		wg.Add(1)
		go startSubscriber(t, i, &wg)
	}

	for i := subscribers + 1; i < publishers+subscribers+1; i++ {
		time.Sleep(time.Millisecond * 20)
		wg.Add(1)
		go startPublisher(t, i, &wg)
	}

	wg.Wait()

	log.Infof("Total Sent %d messages in %d ns, %d ns/msg, %d msgs/sec", totalSent, sentSince, int(float64(sentSince)/float64(totalSent)), int(float64(totalSent)/(float64(sentSince)/float64(time.Second))))
	log.Infof("Total Received %d messages in %d ns, %d ns/msg, %d msgs/sec", totalRcvd, sentSince, int(float64(sentSince)/float64(totalRcvd)), int(float64(totalRcvd)/(float64(sentSince)/float64(time.Second))))
}

func startSubscriber(t testing.TB, cid int, wg *sync.WaitGroup) {
	defer wg.Done()

	now := time.Now()

	clientId := fmt.Sprintf("clientId-%d", cid)
	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID(clientId)
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	cnt := messages * publishers
	received := 0
	since := time.Since(now).Nanoseconds()

	filters := map[string]byte{string(topic): qos}
	token = c.SubscribeMultiple(filters, func(client MQTT.Client, message MQTT.Message) {
		if received == 0 {
			now = time.Now()
		}

		log.Debugf("(cid:%d) messages received=%d", cid, received)
		received++
		since = time.Since(now).Nanoseconds()

		if received == cnt {
			rcvd := atomic.AddInt64(&rcvdone, 1)
			if rcvd == int64(subscribers) {
				close(done2)
			}
		}
	})
	token.WaitTimeout(time.Millisecond * 100)
	token.Wait()
	require.NoError(t, token.Error())

	subs := atomic.AddInt64(&subdone, 1)
	if subs == int64(subscribers) {
		now = time.Now()
		close(done)
	}

	log.Infof("(cid:%d) subscribers ready", cid)

	select {
	case <-done2:
	case <-time.Tick(time.Second * time.Duration(nap*publishers)):
		log.Warnf("(cid:%d) Timed out waiting for messages to be received.", cid)
	}

	token = c.Unsubscribe(string(topic))
	token.Wait()
	token.WaitTimeout(time.Millisecond * 100)
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

	log.Infof("(cid:%d) Received %d messages in %d ns, %d ns/msg, %d msgs/sec", cid, received, since, int(float64(since)/float64(cnt)), int(float64(received)/(float64(since)/float64(time.Second))))
}

func startPublisher(t testing.TB, cid int, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-done:
	case <-time.After(time.Second * time.Duration(subscribers)):
		log.Warnf("(cid:%d) Timed out waiting for subscribe response", cid)
		return
	}

	now := time.Now()

	clientId := fmt.Sprintf("clientId-%d", cid)
	addr := fmt.Sprintf("tcp://%s:%d", host, port)
	opts := MQTT.NewClientOptions().AddBroker(addr).SetClientID(clientId)
	opts.SetDefaultPublishHandler(f)

	c := MQTT.NewClient(opts)
	token := c.Connect()
	token.Wait()
	require.NoError(t, token.Error())

	cnt := messages
	sent := 0

	for i := 0; i < cnt; i++ {
		r := rand.Int63n(200) + 900
		time.Sleep(time.Millisecond * time.Duration(r))

		text := fmt.Sprintf("cid:%d, msg:%d", cid, i)
		token = c.Publish(string(topic), qos, false, []byte(text))
		token.WaitTimeout(time.Millisecond * 100)
		token.Wait()
		require.NoError(t, token.Error())

		sent++
	}

	log.Infof("(cid:%d) publish done", cid)
	since := time.Since(now).Nanoseconds()

	time.Sleep(3 * time.Second)
	c.Disconnect(250)

	statMu.Lock()
	totalSent += int64(sent)
	totalSentTime += int64(since)
	if since > sentSince {
		sentSince = since
	}
	statMu.Unlock()

	log.Infof("(cid:%d) Sent %d messages in %d ns, %d ns/msg, %d msgs/sec", cid, sent, since, int(float64(since)/float64(cnt)), int(float64(sent)/(float64(since)/float64(time.Second))))

	select {
	case <-done2:
	case <-time.Tick(time.Second * time.Duration(nap*publishers)):
		log.Warnf("(cid:%d) Timed out waiting for messages to be received.", cid)
	}
}
