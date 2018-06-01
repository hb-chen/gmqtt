package service

import (
	"net"
	"time"
	"sync"
	"errors"
	"runtime"
	"fmt"

	"github.com/surgemq/message"

	"github.com/hb-go/micro-mq/pkg/log"
	"github.com/hb-go/micro-mq/pkg/gopool"
	"github.com/hb-go/micro-mq/broker"
	"github.com/hb-go/micro-mq/broker/kafka"
	"github.com/hb-go/micro-mq/gateway/sessions"
	"github.com/hb-go/micro-mq/gateway/topics"
	"github.com/hb-go/micro-mq/gateway/auth"
)

var (
	ErrServerClosed = errors.New("server: server closed")
)

var (
	ErrInvalidConnectionType  error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
)

type Server struct {
	ln           net.Listener
	readTimeout  time.Duration
	writeTimeout time.Duration

	// broker
	broker broker.Broker

	// authMgr is the authentication manager that we are going to use for authenticating
	// incoming connections
	authMgr *auth.Manager

	// sessMgr is the sessions manager for keeping track of the sessions
	sessMgr *sessions.Manager

	// topicsMgr is the topics manager for keeping track of subscriptions
	topicMgr *topics.Manager

	//serviceMapMu sync.RWMutex
	//serviceMap   map[string]*service

	mu         sync.RWMutex
	activeConn map[net.Conn]struct{}
	doneChan   chan struct{}

	subs []interface{}
	qoss []byte
}

func NewServer() (srv *Server, err error) {
	srv = &Server{}

	b := kafka.NewBroker()
	if err = b.Connect(); err != nil {
		return nil, err
	} else {
		srv.broker = b
	}

	defer func() {
		if err != nil {
			log.Debugf("server new error:%v", err)
			b.Disconnect()
		}
	}()

	if srv.authMgr, err = auth.NewManager(auth.ProviderNameRpc); err != nil {
		return nil, err
	}

	if srv.sessMgr, err = sessions.NewManager(sessions.ProviderNameMem); err != nil {
		return nil, err
	}

	h := func(p broker.Publication) error {
		return srv.subHandler(p)
	}
	if srv.topicMgr, err = topics.NewManager(topics.ProviderNameMem, srv.broker, h); err != nil {
		return nil, err
	}

	return srv, nil
}

func (s *Server) getDoneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getDoneChanLocked()
}

func (s *Server) getDoneChanLocked() chan struct{} {
	if s.doneChan == nil {
		s.doneChan = make(chan struct{})
	}
	return s.doneChan
}

func (s *Server) closeDoneChanLocked() {
	ch := s.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}

func (srv *Server) ListenAndServe(network, address string) error {
	ln, err := net.Listen(network, address)
	if err != nil {
		log.Errorf("listen and serve error:%v", err)
		return err
	}
	defer ln.Close()

	log.Infof("listen and serve")

	var tempDelay time.Duration

	srv.mu.Lock()
	srv.ln = ln
	if srv.activeConn == nil {
		srv.activeConn = make(map[net.Conn]struct{})
	}
	srv.mu.Unlock()

	connPool := gopool.NewPool(10240, 5, 1)

	for {
		conn, e := ln.Accept()

		if e != nil {
			log.Infof("accept error: %v", e)

			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}

			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Errorf("accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		log.Infof("accept")
		//if tc, ok := conn.(*net.TCPConn); ok {
		//	tc.SetKeepAlive(true)
		//	tc.SetKeepAlivePeriod(30 * time.Second)
		//}

		srv.mu.Lock()
		srv.activeConn[conn] = struct{}{}
		srv.mu.Unlock()

		// @TODO 连接池超出一定等待队列后拒绝，由于是长连接队列应该是0等待，或设计连接数队列等待超时
		err = connPool.Schedule(func() {
			srv.serveConn(conn)
		})
		if err != nil {
			log.Errorf("conn pool schedule error:%v", err)
			conn.Close()
		}
	}
}

func (srv *Server) serveConn(conn net.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			log.Errorf("serving %s panic error: %s, stack:\n %s", conn.RemoteAddr(), err, buf)
		}
		srv.mu.Lock()
		delete(srv.activeConn, conn)
		srv.mu.Unlock()
		conn.Close()

		if err != nil {
			log.Errorf("connection close with err:%v", err)
		}
	}()

	//err = srv.checkConfiguration()
	//if err != nil {
	//	return nil, err
	//}

	//conn, ok := c.(net.Conn)
	//if !ok {
	//	return nil, ErrInvalidConnectionType
	//}

	// To establish a connection, we must
	// 1. Read and decode the message.ConnectMessage from the wire
	// 2. If no decoding errors, then authenticate using username and password.
	//    Otherwise, write out to the wire message.ConnackMessage with
	//    appropriate error.
	// 3. If authentication is successful, then either create a new session or
	//    retrieve existing session
	// 4. Write out to the wire a successful message.ConnackMessage message

	// Read the CONNECT message from the wire, if error, then check to see if it's
	// a CONNACK error. If it's CONNACK error, send the proper CONNACK error back
	// to client. Exit regardless of error type.

	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(2*time.Second)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(conn)
	if err != nil {
		log.Errorf("connect message error:%v", err)
		if cerr, ok := err.(message.ConnackCode); ok {
			log.Debugf("request message: %s\n response message: %s\n error: %v", req, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return err
	}

	log.Infof("serve conn connect message:%v", req)

	// Authenticate the user, if error, return error and exit
	if err = srv.authMgr.Authenticate(string(req.Username()), string(req.Password())); err != nil {
		log.Errorf("authenticate error:%v", err)
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(30)
	}

	svc := &Service{
		conn:     conn,
		broker:   srv.broker,
		topicMgr: srv.topicMgr,
	}

	err = srv.getSession(svc, req, resp)
	if err != nil {
		log.Errorf("serve conn get session error:%v", err)
		return err
	}

	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(conn, resp); err != nil {
		log.Errorf("serve conn connection accepted message error:%v", err)
		return err
	}

	svc.onpub = func(msg *message.PublishMessage) error {
		if err := svc.publish(msg, nil); err != nil {
			log.Errorf("service: publishing message error:%v", err)
			return err
		}

		return nil
	}

	if err := svc.start(); err != nil {
		log.Errorf("serve conn service error:%v", err)
		svc.stop()
		return err
	}

	return nil
}

func (this *Server) subHandler(p broker.Publication) error {
	log.Debug("[sub] received message:", string(p.Message().Body), "header", p.Message().Header)

	// @TODO msg Encode/Decode
	var qos, topic, ok = "", "", false
	if qos, ok = p.Message().Header[topics.MQHeaderMQTTQos]; !ok {
		return errors.New("broker msg header error:qos nil")
	}

	if topic, ok = p.Message().Header[topics.MQHeaderMQTTTopic]; !ok {
		return errors.New("broker msg header error:topic nil")
	}
	msg := message.NewPublishMessage()
	msg.SetTopic([]byte(topic))
	msg.SetQoS(byte(qos[0]))
	msg.SetPayload(p.Message().Body)

	if msg.Retain() {
		if err := this.topicMgr.Retain(msg); err != nil {
			log.Errorf("Error retaining message: %v", err)
		}
	}

	if err := this.topicMgr.Subscribers(msg.Topic(), msg.QoS(), &this.subs, &this.qoss); err != nil {
		return err
	}

	msg.SetRetain(false)

	//log.Debugf("(%s) Publishing to topic %q and %d subscribers", this.cid(), string(msg.Topic()), len(this.subs))
	for _, s := range this.subs {
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

func (this *Server) getSession(svc *Service, req *message.ConnectMessage, resp *message.ConnackMessage) error {
	// If CleanSession is set to 0, the server MUST resume communications with the
	// client based on state from the current session, as identified by the client
	// identifier. If there is no session associated with the client identifier the
	// server must create a new session.
	//
	// If CleanSession is set to 1, the client and server must discard any previous
	// session and start a new one. This session lasts as long as the network c
	// onnection. State data associated with this session must not be reused in any
	// subsequent session.

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient%d", svc.id)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if svc.sess, err = this.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := svc.sess.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if svc.sess == nil {
		if svc.sess, err = this.sessMgr.New(cid); err != nil {
			return err
		}

		resp.SetSessionPresent(false)

		if err := svc.sess.Init(req); err != nil {
			return err
		}
	}

	return nil
}
