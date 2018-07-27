package server

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/recallsong/go-utils/errorx"
	"github.com/recallsong/go-utils/net/echox"
	"github.com/recallsong/go-utils/net/servegrp"
	"github.com/recallsong/online-statistics/server/store"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	cfg          *Config
	svr          *echox.EchoServer
	admin_svr    *echox.EchoServer
	serGrp       *servegrp.ServeGroup
	watchers     map[string]*websocket.Conn
	watchersLock sync.RWMutex
	conns        map[string]*Conn
	connsLock    sync.RWMutex
	connNum      int32
	store        store.Store
}

func New(cfg *Config) *Server {
	cfg.HttpAddr = strings.TrimSpace(cfg.HttpAddr)
	cfg.HttpsAddr = strings.TrimSpace(cfg.HttpsAddr)
	cfg.AdminAddr = strings.TrimSpace(cfg.AdminAddr)
	return &Server{
		cfg:      cfg,
		serGrp:   servegrp.NewServeGroup(),
		svr:      echox.New(),
		conns:    make(map[string]*Conn),
		watchers: make(map[string]*websocket.Conn),
		store:    newStore(cfg.Store),
	}
}

func (s *Server) setEchoLog(svr *echo.Echo) {
}

func (s *Server) Start(closeCh <-chan os.Signal) error {
	s.setEchoLog(s.svr.Echo)
	if s.cfg.HttpAddr != "" {
		_, svr := s.svr.GetHttpServer(s.cfg.HttpAddr)
		err := s.serGrp.Put(s.cfg.HttpAddr, svr)
		if err != nil {
			log.Errorf("[server] %v", err)
			return err
		}
	}
	if s.cfg.HttpsAddr != "" {
		parts := strings.Split(s.cfg.HttpsAddr, ",")
		if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
			err := fmt.Errorf("https address format is invalid")
			log.Error("[server] ", err)
			return err
		}
		_, svr := s.svr.GetHttpsServer(parts[0], parts[1], parts[2])
		err := s.serGrp.Put(parts[0], svr)
		if err != nil {
			log.Errorf("[server] %v", err)
			return err
		}
	}

	if s.cfg.TcpAddr != "" {
		// TODO
	}
	if s.cfg.TcpTLSAddr != "" {
		// TODO
	}

	if s.serGrp.Num() <= 0 {
		err := errorx.New("no address to listen")
		log.Errorf("[server] %v", err)
		return err
	}
	s.initHttpRoutes(s.svr)

	if s.cfg.AdminAddr != "" {
		s.admin_svr = echox.New()
		s.setEchoLog(s.admin_svr.Echo)
		_, svr := s.admin_svr.GetHttpServer(s.cfg.AdminAddr)
		err := s.serGrp.Put(s.cfg.AdminAddr, svr)
		if err != nil {
			log.Errorf("[server] %v", err)
			return err
		}
		s.initAdminRoutes(s.admin_svr)
	}

	err := s.serGrp.Serve(closeCh, func(err error, addr string, svr servegrp.ServeItem) {
		if err != nil {
			log.Errorf("[server] close [ %s ] error: %v", addr, err)
		} else {
			log.Infof("[server] close [ %s ] ok", addr)
		}
	})
	if err != nil {
		log.Error("[server] close server with errors")
	} else {
		log.Infof("[server] close server ok")
	}
	return err
}
