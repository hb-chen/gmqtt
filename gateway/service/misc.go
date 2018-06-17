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

package service

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/surgemq/message"

	"github.com/hb-go/micro-mq/pkg/log"
)

func getConnectMessage(conn io.Reader) (*message.ConnectMessage, error) {
	buf, err := getMessageBuffer(conn)
	if err != nil {
		log.Infof("Receive error: %v", err)
		return nil, err
	}

	msg := message.NewConnectMessage()

	_, err = msg.Decode(buf)
	log.Debugf("Received: %s", msg)
	return msg, err
}

func getConnackMessage(conn io.Reader) (*message.ConnackMessage, error) {
	buf, err := getMessageBuffer(conn)
	if err != nil {
		log.Infof("Receive error: %v", err)
		return nil, err
	}

	msg := message.NewConnackMessage()

	_, err = msg.Decode(buf)
	log.Debugf("Received: %s", msg)
	return msg, err
}

func writeMessage(conn io.Writer, msg message.Message) error {
	buf := make([]byte, msg.Len())
	_, err := msg.Encode(buf)
	if err != nil {
		log.Infof("Write error: %v", err)
		return err
	}
	log.Debugf("Writing: %s", msg)

	return writeMessageBuffer(conn, buf)
}

func getMessageBuffer(c io.Reader) ([]byte, error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}
	//conn, ok := c.(net.Conn)
	//if !ok {
	//	return nil, ErrInvalidConnectionType
	//}

	var (
		// the message buffer
		buf []byte

		// tmp buffer to read a single byte
		b []byte = make([]byte, 1)

		// total bytes read
		l int = 0
	)

	// Let's read enough bytes to get the message header (msg type, remaining length)
	for {
		// If we have read 5 bytes and still not done, then there's a problem.
		if l > 5 {
			return nil, fmt.Errorf("connect/getMessage: 4th byte of remaining length has continuation bit set")
		}

		n, err := c.Read(b[0:])
		if err != nil {
			//log.Debugf("Read error: %v", err)
			return nil, err
		}

		// Technically i don't think we will ever get here
		if n == 0 {
			continue
		}

		buf = append(buf, b...)
		l += n

		// Check the remlen byte (1+) to see if the continuation bit is set. If so,
		// increment cnt and continue reading. Otherwise break.
		if l > 1 && b[0] < 0x80 {
			break
		}
	}

	// Get the remaining length of the message
	remlen, _ := binary.Uvarint(buf[1:])
	buf = append(buf, make([]byte, remlen)...)

	for l < len(buf) {
		n, err := c.Read(buf[l:])
		if err != nil {
			return nil, err
		}
		l += n
	}

	return buf, nil
}

func writeMessageBuffer(c io.Writer, b []byte) error {
	if c == nil {
		return ErrInvalidConnectionType
	}

	//conn, ok := c.(net.Conn)
	//if !ok {
	//	return ErrInvalidConnectionType
	//}

	_, err := c.Write(b)
	return err
}

// Copied from http://golang.org/src/pkg/net/timeout_test.go
func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}
