package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/recallsong/go-utils/container/dic"
	"github.com/recallsong/online-statistics/server/store"
)

type RedisStore struct {
	pool          *redis.Pool
	pushEvents    bool
	eventsKey     string
	saveList      bool
	listKeyPrefix string
}

func NewStore(options map[string]interface{}) (store.Store, error) {
	cfg := dic.FromMap(options)
	host := cfg.GetString("host", "")
	if len(host) <= 0 {
		return nil, nil
	}
	eventsKey := cfg.GetString("events_key", "")
	listKeyPrefix := cfg.GetString("list_key_prefix", "")
	if len(eventsKey) <= 0 && len(listKeyPrefix) <= 0 {
		return nil, fmt.Errorf("store events_key or list_key_prefix are set to at least one")
	}
	connTime := cfg.GetDuration("connect_timeout", 3*time.Second)
	readTime := cfg.GetDuration("read_timeout", 3*time.Second)
	writeTime := cfg.GetDuration("write_timeout", 3*time.Second)
	password := cfg.GetString("password", "")
	dbIndex := cfg.GetInt("db", 0)
	pool := &redis.Pool{
		MaxIdle:     cfg.GetInt("max_idle", 3),
		IdleTimeout: cfg.GetDuration("idle_timeout", 240*time.Second),
		Wait:        true,
		MaxActive:   cfg.GetInt("max_active", 30),
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", host, connTime, readTime, writeTime)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", dbIndex); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: nil,
	}
	return &RedisStore{
		pool:          pool,
		pushEvents:    eventsKey != "",
		eventsKey:     eventsKey,
		saveList:      listKeyPrefix != "",
		listKeyPrefix: listKeyPrefix,
	}, nil
}

func (rs *RedisStore) Online(evt *store.Event) error {
	conn := rs.pool.Get()
	defer conn.Close()
	if rs.pushEvents {
		bytes, _ := json.Marshal(evt)
		_, err := conn.Do("LPUSH", rs.eventsKey, bytes)
		if err != nil {
			return err
		}
	}
	if rs.saveList {
		_, err := conn.Do("ZADD", rs.listKeyPrefix+evt.Topic, time.Now().Unix(), evt.Addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *RedisStore) Offline(evt *store.Event) error {
	conn := rs.pool.Get()
	defer conn.Close()
	if rs.pushEvents {
		bytes, _ := json.Marshal(evt)
		_, err := conn.Do("LPUSH", rs.eventsKey, bytes)
		if err != nil {
			return err
		}
	}
	if rs.saveList {
		_, err := conn.Do("ZREM", rs.listKeyPrefix+evt.Topic, evt.Addr)
		if err != nil {
			return err
		}
	}
	return nil
}
