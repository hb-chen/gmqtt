package store

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/hb-go/micro-mq/gateway/sessions"
)

type MockStore struct {
	st map[string]*sessions.Session
	mu sync.RWMutex
}

func NewMockStore() (*MockStore, error) {
	mock := &MockStore{
		st: make(map[string]*sessions.Session),
	}

	return mock, nil
}

func (this *MockStore) Get(id string) (*sessions.Session, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()

	sess, ok := this.st[id]
	if !ok {
		return nil, fmt.Errorf("store/Get: No session found for key %s", id)
	}

	return sess, nil
}

func (this *MockStore) Set(id string, sess *sessions.Session) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.st[id] = sess
	return nil
}

func (this *MockStore) Del(id string) error {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.st, id)
	return nil
}

func (this *MockStore) Range(page, size int64) ([]*sessions.Session, error) {
	keys := reflect.ValueOf(this.st).MapKeys()

	start := page * size
	stop := start + size
	l := int64(len(keys))
	if stop > l {
		stop = l
	}

	sesses := make([]*sessions.Session, start-stop)
	for _, k := range keys {
		sesses = append(sesses, this.st[k.String()])
	}

	return sesses, nil
}

func (this *MockStore) Count() (int64, error) {
	return int64(len(this.st)), nil
}

func (this *MockStore) Close() error {
	this.st = make(map[string]*sessions.Session)
	return nil
}
