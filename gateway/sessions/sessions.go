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

package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrSessionsProviderNotFound = errors.New("Session: Session provider not found")
	ErrKeyNotAvailable          = errors.New("Session: not item found for key.")

	providers = make(map[string]SessionsProvider)
)

type SessionsProvider interface {
	New(id string) (*Session, error)
	Get(id string) (*Session, error)
	Del(id string)
	Save(id string) error
	Count() int
	Close() error
}

// Register makes a session provider available by the provided name.
// If a Register is called twice with the same name or if the driver is nil,
// it panics.
func Register(name string, provider SessionsProvider) {
	if provider == nil {
		panic("session: Register provide is nil")
	}

	if _, dup := providers[name]; dup {
		panic("session: Register called twice for provider " + name)
	}

	providers[name] = provider
}

func Unregister(name string) {
	delete(providers, name)
}

type Manager struct {
	store Store
}

func NewManager(store Store) (*Manager) {
	return &Manager{store: store}
}

func (this *Manager) New(id string) (*Session, error) {
	if id == "" {
		id = this.sessionId()
	}

	sess := &Session{id: id}
	err := this.store.Set(id, sess)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (this *Manager) Get(id string) (*Session, error) {
	return this.store.Get(id)
}

func (this *Manager) Del(id string) error {
	return this.store.Del(id)
}

func (this *Manager) Save(id string, sess *Session) error {
	return this.store.Set(id, sess)
}

func (this *Manager) Count() (int64, error) {
	return this.store.Count()
}

func (this *Manager) Close() error {
	return this.store.Close()
}

func (manager *Manager) sessionId() string {
	b := make([]byte, 15)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
