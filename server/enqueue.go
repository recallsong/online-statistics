package server

import (
	"github.com/recallsong/go-utils/container/dic"
	"github.com/recallsong/online-statistics/server/store"
	redis "github.com/recallsong/online-statistics/server/store/redis-store"
	log "github.com/sirupsen/logrus"
)

func newStore(cfg map[string]interface{}) store.Store {
	if cfg == nil {
		return nil
	}
	dic := dic.FromMap(cfg)
	if dic.GetString("key", "redis") != "redis" {
		log.Fatal("invalid store name")
		return nil
	}
	s, err := redis.NewStore(cfg)
	if err != nil {
		log.Fatal("create store error : ", err)
		return nil
	}
	return s
}

func (s *Server) pushOnlineEvent(addr string, conn *Conn) error {
	err := s.store.Online(&store.Event{
		Action:  "online",
		StartOn: conn.StartOn.UnixNano() / 1000000,
		Topic:   conn.Pkg.Topic,
		Token:   conn.Pkg.Token,
		Domain:  conn.Pkg.Domain,
	})
	if err != nil {
		log.Error("[queue] push online event to store error : ", err)
		return err
	}
	return nil
}

func (s *Server) pushOfflineEvent(addr string, conn *Conn) error {
	err := s.store.Offline(&store.Event{
		Action:  "offline",
		StartOn: conn.StartOn.UnixNano() / 1000000,
		Topic:   conn.Pkg.Topic,
		Token:   conn.Pkg.Token,
		Domain:  conn.Pkg.Domain,
	})
	if err != nil {
		log.Error("[queue] push offline event to store error : ", err)
		return err
	}
	return nil
}
