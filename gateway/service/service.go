package service

import (
	"net"
	"bufio"
	"sync"
	"io"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/mailru/easygo/netpoll"
	"github.com/surgemq/message"

	"github.com/hb-go/micro-mq/pkg/log"
	"github.com/hb-go/micro-mq/broker"
	"github.com/hb-go/micro-mq/gateway/sessions"
	"github.com/hb-go/micro-mq/gateway/topics"
)

type (
	OnCompleteFunc func(msg, ack message.Message, err error) error
	OnPublishFunc func(msg *message.PublishMessage) error
)

var (
	poller netpoll.Poller
	mu     sync.Mutex
	events []netpoll.Event
)

var (
	errDisconnect = errors.New("Disconnect")
)

func init() {
	var err error
	poller, err = netpoll.New(&netpoll.Config{})
	if err != nil {
		panic(err)
	}
}

type service struct {
	id   int64
	conn net.Conn

	// broker
	broker broker.Broker

	sess *sessions.Session

	sessMgr  *sessions.Manager
	topicMgr *topics.Manager

	wgStarted sync.WaitGroup
	wgStopped sync.WaitGroup

	onpub OnPublishFunc

	subs  []interface{}
	qoss  []byte
	rmsgs []*message.PublishMessage

	// Whether this is service is closed or not.
	closed int64
}

func (svc *service) start() (err error) {
	defer func() {
		svc.stop()
	}()

	// If this is a recovered session, then add any topics it subscribed before
	topics, qoss, err := svc.sess.Topics()
	if err != nil {
		return err
	} else {
		for i, t := range topics {
			svc.topicMgr.Subscribe([]byte(t), qoss[i], &svc.onpub)
		}
	}

	// netpoll实现
	desc := netpoll.Must(netpoll.HandleRead(svc.conn))
	defer func() {
		if err = poller.Stop(desc); err != nil {
			log.Errorf("poller stop error: %v", err)
		}
	}()

	exit := make(chan bool, 1)
	var (
		data     = []byte("hello")
		received = make([]byte, 0, len(data))
	)

	err = poller.Start(desc, func(event netpoll.Event) {
		mu.Lock()
		events = append(events, event)
		mu.Unlock()

		log.Infof("poller event:%v", event)
		if event&netpoll.EventRead == 0 {
			return
		}

		bts := make([]byte, 128)
		//n, err := conn.Read(bts)
		buf := bufio.NewReader(svc.conn)
		n, err := buf.Read(bts)
		switch {
		case err != nil:
			log.Errorf("poller read error:%v", err)

			if err == io.EOF {
				exit <- true
			}

			return
		default:
			log.Debugf("(%s) poller received:%v", svc.sess.ID(), bts)
			received = append(received, bts[:n]...)
			//if err := poller.Resume(desc); err != nil {
			//	log.Errorf("poller.Resume() error: %v", err)
			//}

			mtype := message.MessageType(bts[0] >> 4)

			msg, err := mtype.New()
			if err != nil {
				log.Errorf("message type new error:%v", err)
				return
			}

			n, err = msg.Decode(bts)
			if err != nil {
				log.Errorf("message decode error:%v", err)
				return
			} else {

			}

			log.Infof("receive message:%v", msg)

			err = svc.process(msg)
			if err != nil {
				log.Errorf("message process error:%v", err)
				if err == errDisconnect {
					exit <- true
				}
			}
		}

		received = received[0:0]
	})

	// Receiver is responsible for reading from the connection and putting data into
	// a buffer.
	//svc.wgStarted.Add(1)
	//svc.wgStopped.Add(1)
	//go svc.receiver()

	<-exit
	log.Infof("service exit")

	return nil
}

func (svc *service) stop() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			log.Errorf("(%s) Recovering from panic: %v", svc.id, r)
		}
	}()

	doit := atomic.CompareAndSwapInt64(&svc.closed, 0, 1)
	if !doit {
		return
	}

	log.Debugf("service stop")

	// Close quit channel, effectively telling all the goroutines it's time to quit
	//if svc.done != nil {
	//	log.Debugf("(%s) closing this.done", svc.id)
	//	close(svc.done)
	//}

	// Close the network connection
	if svc.conn != nil {
		log.Debugf("(%s) closing this.conn", svc.id)
		svc.conn.Close()
	}

	// Wait for all the goroutines to stop.
	//svc.wgStopped.Wait()

	//log.Debugf("(%s) Received %d bytes in %d messages.", this.cid(), this.inStat.bytes, this.inStat.msgs)
	//log.Debugf("(%s) Sent %d bytes in %d messages.", this.cid(), this.outStat.bytes, this.outStat.msgs)

	// Unsubscribe from all the topics for this client, only for the server side though
	if svc.sess != nil {
		topics, _, err := svc.sess.Topics()
		if err != nil {
			log.Errorf("(%s/%d): %v", svc.id, svc.id, err)
		} else {
			for _, t := range topics {
				if err := svc.topicMgr.Unsubscribe([]byte(t), &svc.onpub); err != nil {
					log.Errorf("(%s): Error unsubscribing topic %q: %v", svc.id, t, err)
				}
			}
		}
	}

	// Publish will message if WillFlag is set. Server side only.
	if svc.sess.Cmsg.WillFlag() && svc.sess.Will != nil {
		log.Infof("(%s) service/stop: connection unexpectedly closed. Sending Will.", svc.id)
		svc.onPublish(svc.sess.Will)
	}

	// Remove the client topics manager
	//if svc.client {
	//	topics.Unregister(svc.sess.ID())
	//}

	// Remove the session from session store if it's suppose to be clean session
	if svc.sess.Cmsg.CleanSession() && svc.sessMgr != nil {
		svc.sessMgr.Del(svc.sess.ID())
	} else {
		svc.sessMgr.Save(svc.sess.ID())
	}

	svc.conn = nil
}

func (svc *service) process(msg message.Message) error {
	log.Debugf("process msg:%v", msg)
	var err error = nil

	switch msg := msg.(type) {
	case *message.PublishMessage:
		// For PUBLISH message, we should figure out what QoS it is and process accordingly
		// If QoS == 0, we should just take the next step, no ack required
		// If QoS == 1, we should send back PUBACK, then take the next step
		// If QoS == 2, we need to put it in the ack queue, send back PUBREC
		err = svc.processPublish(msg)

	case *message.PubackMessage:
		// For PUBACK message, it means QoS 1, we should send to ack queue
		svc.sess.Pub1ack.Ack(msg)
		svc.processAcked(svc.sess.Pub1ack)

	case *message.PubrecMessage:
		// For PUBREC message, it means QoS 2, we should send to ack queue, and send back PUBREL
		if err = svc.sess.Pub2out.Ack(msg); err != nil {
			break
		}

		resp := message.NewPubrelMessage()
		resp.SetPacketId(msg.PacketId())
		_, err = svc.writeMessage(resp)

	case *message.PubrelMessage:
		// For PUBREL message, it means QoS 2, we should send to ack queue, and send back PUBCOMP
		if err = svc.sess.Pub2in.Ack(msg); err != nil {
			break
		}

		svc.processAcked(svc.sess.Pub2in)

		resp := message.NewPubcompMessage()
		resp.SetPacketId(msg.PacketId())
		_, err = svc.writeMessage(resp)

	case *message.PubcompMessage:
		// For PUBCOMP message, it means QoS 2, we should send to ack queue
		if err = svc.sess.Pub2out.Ack(msg); err != nil {
			break
		}

		svc.processAcked(svc.sess.Pub2out)

	case *message.SubscribeMessage:
		// For SUBSCRIBE message, we should add subscriber, then send back SUBACK
		return svc.processSubscribe(msg)

	case *message.SubackMessage:
		// For SUBACK message, we should send to ack queue
		svc.sess.Suback.Ack(msg)
		svc.processAcked(svc.sess.Suback)

	case *message.UnsubscribeMessage:
		// For UNSUBSCRIBE message, we should remove subscriber, then send back UNSUBACK
		return svc.processUnsubscribe(msg)

	case *message.UnsubackMessage:
		// For UNSUBACK message, we should send to ack queue
		svc.sess.Unsuback.Ack(msg)
		svc.processAcked(svc.sess.Unsuback)

	case *message.PingreqMessage:
		// For PINGREQ message, we should send back PINGRESP
		resp := message.NewPingrespMessage()
		_, err = svc.writeMessage(resp)

	case *message.PingrespMessage:
		svc.sess.Pingack.Ack(msg)
		svc.processAcked(svc.sess.Pingack)

	case *message.DisconnectMessage:
		// For DISCONNECT message, we should quit
		svc.sess.Cmsg.SetWillFlag(false)
		return errDisconnect

	default:
		return fmt.Errorf("(%d) invalid message type %s.", svc.id, msg.Name())
	}

	if err != nil {
		log.Debugf("(%s) Error processing acked message: %v", svc.id, err)
	}

	return err
}

func (svc *service) processPublish(msg *message.PublishMessage) error {
	log.Debugf("process publish msg:%v", msg)

	switch msg.QoS() {
	case message.QosExactlyOnce:
		svc.sess.Pub2in.Wait(msg, nil)

		resp := message.NewPubrecMessage()
		resp.SetPacketId(msg.PacketId())

		_, err := svc.writeMessage(resp)
		return err

	case message.QosAtLeastOnce:
		resp := message.NewPubackMessage()
		resp.SetPacketId(msg.PacketId())

		if _, err := svc.writeMessage(resp); err != nil {
			return err
		}

		return svc.onPublish(msg)

	case message.QosAtMostOnce:
		return svc.onPublish(msg)
	}

	return fmt.Errorf("(%d) invalid message QoS %d.", svc.id, msg.QoS())
}

func (svc *service) processAcked(ackq *sessions.Ackqueue) {
	log.Debugf("process acked")

	for _, ackmsg := range ackq.Acked() {
		// Let's get the messages from the saved message byte slices.
		msg, err := ackmsg.Mtype.New()
		if err != nil {
			log.Errorf("process/processAcked: Unable to creating new %s message: %v", ackmsg.Mtype, err)
			continue
		}

		if _, err := msg.Decode(ackmsg.Msgbuf); err != nil {
			log.Errorf("process/processAcked: Unable to decode %s message: %v", ackmsg.Mtype, err)
			continue
		}

		ack, err := ackmsg.State.New()
		if err != nil {
			log.Errorf("process/processAcked: Unable to creating new %s message: %v", ackmsg.State, err)
			continue
		}

		if _, err := ack.Decode(ackmsg.Ackbuf); err != nil {
			log.Errorf("process/processAcked: Unable to decode %s message: %v", ackmsg.State, err)
			continue
		}

		//log.Debugf("(%s) Processing acked message: %v", svc.cid(), ack)

		// - PUBACK if it's QoS 1 message. This is on the client side.
		// - PUBREL if it's QoS 2 message. This is on the server side.
		// - PUBCOMP if it's QoS 2 message. This is on the client side.
		// - SUBACK if it's a subscribe message. This is on the client side.
		// - UNSUBACK if it's a unsubscribe message. This is on the client side.
		switch ackmsg.State {
		case message.PUBREL:
			// If ack is PUBREL, that means the QoS 2 message sent by a remote client is
			// releassed, so let's publish it to other subscribers.
			if err = svc.onPublish(msg.(*message.PublishMessage)); err != nil {
				log.Errorf("(%d) Error processing ack'ed %s message: %v", svc.id, ackmsg.Mtype, err)
			}

		case message.PUBACK, message.PUBCOMP, message.SUBACK, message.UNSUBACK, message.PINGRESP:
			log.Debugf("process/processAcked: %s", ack)
			// If ack is PUBACK, that means the QoS 1 message sent by this service got
			// ack'ed. There's nothing to do other than calling onComplete() below.

			// If ack is PUBCOMP, that means the QoS 2 message sent by this service got
			// ack'ed. There's nothing to do other than calling onComplete() below.

			// If ack is SUBACK, that means the SUBSCRIBE message sent by this service
			// got ack'ed. There's nothing to do other than calling onComplete() below.

			// If ack is UNSUBACK, that means the SUBSCRIBE message sent by this service
			// got ack'ed. There's nothing to do other than calling onComplete() below.

			// If ack is PINGRESP, that means the PINGREQ message sent by this service
			// got ack'ed. There's nothing to do other than calling onComplete() below.

			err = nil

		default:
			log.Errorf("(%d) Invalid ack message type %s.", svc.id, ackmsg.State)
			continue
		}

		// Call the registered onComplete function
		if ackmsg.OnComplete != nil {
			onComplete, ok := ackmsg.OnComplete.(OnCompleteFunc)
			if !ok {
				log.Errorf("process/processAcked: Error type asserting onComplete function: %v", reflect.TypeOf(ackmsg.OnComplete))
			} else if onComplete != nil {
				if err := onComplete(msg, ack, nil); err != nil {
					log.Errorf("process/processAcked: Error running onComplete(): %v", err)
				}
			}
		}
	}
}

func (svc *service) processSubscribe(msg *message.SubscribeMessage) error {
	log.Debugf("process subscribe msg:%v", msg)
	resp := message.NewSubackMessage()
	resp.SetPacketId(msg.PacketId())

	// Subscribe to the different topics
	var retcodes []byte

	topics := msg.Topics()
	qos := msg.Qos()

	svc.rmsgs = svc.rmsgs[0:0]

	for i, t := range topics {
		rqos, err := svc.topicMgr.Subscribe(t, qos[i], &svc.onpub)
		if err != nil {
			return err
		}
		svc.sess.AddTopic(string(t), qos[i])

		retcodes = append(retcodes, rqos)

		// yeah I am not checking errors here. If there's an error we don't want the
		// subscription to stop, just let it go.
		svc.topicMgr.Retained(t, &svc.rmsgs)
		log.Debugf("(%d) topic = %s, retained count = %d", svc.id, string(t), len(svc.rmsgs))
	}

	if err := resp.AddReturnCodes(retcodes); err != nil {
		return err
	}

	if _, err := svc.writeMessage(resp); err != nil {
		return err
	}

	for _, rm := range svc.rmsgs {
		if err := svc.publish(rm, nil); err != nil {
			log.Errorf("service/processSubscribe: Error publishing retained message: %v", err)
			return err
		}
	}

	return nil
}

func (svc *service) processUnsubscribe(msg *message.UnsubscribeMessage) error {
	log.Debugf("process unsubscribe msg:%v", msg)
	topics := msg.Topics()

	for _, t := range topics {
		svc.topicMgr.Unsubscribe(t, &svc.onpub)
		svc.sess.RemoveTopic(string(t))
	}

	resp := message.NewUnsubackMessage()
	resp.SetPacketId(msg.PacketId())

	_, err := svc.writeMessage(resp)
	return err
	return nil
}

// writeMessage() writes a message to the outgoing buffer
func (svc *service) writeMessage(msg message.Message) (int, error) {
	err := writeMessage(svc.conn, msg)
	return 0, err
}

func (svc *service) onPublish(msg *message.PublishMessage) error {
	//log.Errorf("client:(%s) on publish message: %v", svc.cid(), msg.String())

	// 发布消息走broker
	if svc.broker != nil {
		// @TODO Message Header加入MQTT特有属性 QoS、Topic
		header := map[string]string{
			topics.MQHeaderMQTTQos:   string(msg.QoS()),
			topics.MQHeaderMQTTTopic: string(msg.Topic()),
		}
		bMsg := broker.Message{
			Header: header,
			Body:   msg.Payload(),
		}

		err := svc.broker.Publish(topics.TopicToBrokerTopic(msg.Topic()), &bMsg)
		if err != nil {
			return err
		}

		return nil
	}

	if msg.Retain() {
		if err := svc.topicMgr.Retain(msg); err != nil {
			log.Errorf("(%d) Error retaining message: %v", svc.id, err)
		}
	}

	err := svc.topicMgr.Subscribers(msg.Topic(), msg.QoS(), &svc.subs, &svc.qoss)
	if err != nil {
		log.Errorf("(%d) Error retrieving subscribers list: %v", svc.id, err)
		return err
	}

	msg.SetRetain(false)

	//log.Debugf("(%s) Publishing to topic %q and %d subscribers", svc.cid(), string(msg.Topic()), len(svc.subs))
	for _, s := range svc.subs {
		if s != nil {
			fn, ok := s.(*OnPublishFunc)
			if !ok {
				log.Errorf("Invalid onPublish Function")
				return fmt.Errorf("Invalid onPublish Function")
			} else {
				(*fn)(msg)
			}
		}
	}

	return nil
}

func (svc *service) publish(msg *message.PublishMessage, onComplete OnCompleteFunc) error {
	log.Debugf("service/publish: Publishing %s", msg)

	_, err := svc.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("sending message error:%v", err)
	}

	switch msg.QoS() {
	case message.QosAtMostOnce:
		if onComplete != nil {
			return onComplete(msg, nil, nil)
		}

		return nil

	case message.QosAtLeastOnce:
		return svc.sess.Pub1ack.Wait(msg, onComplete)

	case message.QosExactlyOnce:
		return svc.sess.Pub2out.Wait(msg, onComplete)
	}

	return nil
}
