package store

import (
	"time"

	"github.com/go-redis/redis"

	"github.com/hb-chen/gmqtt/gateway/sessions"
)

const sorted_set_key = "sorted_set_key"

type RedisStore struct {
	client *redis.Client
}

// NewRedisStore("127.0.0.1:6379","123456")
func NewRedisStore(addr string, password string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, nil
	}

	rs := &RedisStore{
		client: client,
	}

	return rs, nil
}

func (this *RedisStore) Get(id string) (*sessions.Session, error) {
	val, err := this.client.Get(id).Bytes()

	sess := &sessions.Session{}
	err = sess.Deserialize(val)
	if err != nil {
		return nil, err
	}

	return sess, err
}

func (this *RedisStore) Set(id string, sess *sessions.Session) error {
	b, err := sess.Serialize()
	if err != nil {
		return err
	}

	// @TODO 是否需要事务
	z := redis.Z{Score: float64(time.Now().Nanosecond()), Member: id}
	err = this.client.ZAdd(sorted_set_key, z).Err()
	if err != nil {
		return err
	}

	return this.client.Set(id, b, 0).Err()
}

func (this *RedisStore) Del(id string) error {
	// @TODO 是否需要事务
	err := this.client.ZRem(sorted_set_key, id).Err()
	if err != nil {
		return err
	}

	return this.client.Del(id).Err()
}

func (this *RedisStore) Range(page, size int64) ([]*sessions.Session, error) {
	start := page * size
	stop := start + size
	vals, err := this.client.ZRange(sorted_set_key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	sesses := make([]*sessions.Session, len(vals))
	for _, val := range vals {
		sess := &sessions.Session{}
		err = sess.Deserialize([]byte(val))
		if err != nil {
			return nil, err
		}

		sesses = append(sesses, sess)
	}

	return sesses, nil
}

func (this *RedisStore) Count() (int64, error) {
	return this.client.ZCount("key", "-inf", "+inf").Result()
}

func (this *RedisStore) Close() error {
	return this.client.Close()
}
