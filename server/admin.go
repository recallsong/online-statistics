package server

import (
	"encoding/json"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/recallsong/go-utils/container/dic"
	"github.com/recallsong/go-utils/errorx"
	"github.com/recallsong/go-utils/net/echox"
	"github.com/recallsong/go-utils/reflectx"
	log "github.com/sirupsen/logrus"
)

func (s *Server) initAdminRoutes(svr *echox.EchoServer) {
	svr.Static("/admin", "page")
	svr.GET("/api/clients/ws", s.httpWsListClients)
	svr.GET("/api/clients", s.httpShowClients)
	svr.GET("/api/clients/num", s.httpCloseClient)
	svr.DELETE("/api/clients", s.httpCloseClient)
}

type connInfo struct {
	Addr    string `json:"addr"`
	Topic   string `json:"topic"`
	Token   string `json:"token"`
	Domain  string `json:"domain"`
	StartOn int64  `json:"startOn"`
}

func (s *Server) httpShowClients(c echo.Context) error {
	s.connsLock.RLock()
	list := make([]*connInfo, len(s.conns), len(s.conns))
	i := 0
	for k, v := range s.conns {
		list[i] = &connInfo{
			Addr:    k,
			Topic:   v.Pkg.Topic,
			Token:   v.Pkg.Token,
			Domain:  v.Pkg.Domain,
			StartOn: v.StartOn.UnixNano() / 1000000,
		}
		i++
	}
	s.connsLock.RUnlock()
	return c.JSON(200, dic.Dic{
		"data": dic.Dic{
			"total": len(list),
			"list":  list,
		},
	})
}

func (s *Server) httpShowClientsNum(c echo.Context) error {
	return c.JSON(200, dic.Dic{
		"data": atomic.LoadInt32(&s.connNum),
	})
}

func (s *Server) httpCloseClient(c echo.Context) error {
	addr := c.QueryParam("addr")
	if len(addr) <= 0 {
		return c.JSON(400, dic.Dic{
			"message": "addr should not be empty",
		})
	}
	s.connsLock.RLock()
	if conn, ok := s.conns[addr]; ok {
		s.connsLock.RUnlock()
		err := conn.Conn.Close()
		if err != nil {
			return c.JSON(500, dic.Dic{
				"message": "close " + addr + " not exist",
			})
		}
		return c.JSON(200, dic.Dic{
			"data": connInfo{
				Addr:    addr,
				Topic:   conn.Pkg.Topic,
				Token:   conn.Pkg.Token,
				Domain:  conn.Pkg.Domain,
				StartOn: conn.StartOn.UnixNano() / 1000000,
			},
		})
	} else {
		s.connsLock.RUnlock()
		return c.JSON(400, dic.Dic{
			"message": addr + " not exist",
		})
	}
}

func (s *Server) notifyToAdminPage(addr string, conn *Conn, atcion string) {
	total := atomic.LoadInt32(&s.connNum)
	msg, _ := json.Marshal(dic.Dic{
		"total":  total,
		"action": atcion,
		"item": &connInfo{
			Addr:    addr,
			Topic:   conn.Pkg.Topic,
			Token:   conn.Pkg.Token,
			Domain:  conn.Pkg.Domain,
			StartOn: conn.StartOn.UnixNano() / 1000000,
		},
	})
	s.watchersLock.RLock()
	for _, v := range s.watchers {
		v.WriteMessage(websocket.TextMessage, msg)
	}
	s.watchersLock.RUnlock()
}

func (s *Server) httpWsListClients(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	mtyp, msg, err := ws.ReadMessage()
	if err != nil {
		log.Error("[admin] [ws] read error: ", err)
		return err
	}
	if mtyp != websocket.TextMessage {
		log.Error("[admin] ws read a non TextMessage")
		return errorx.New("not text message")
	}
	if reflectx.BytesToString(msg) != "watch" {
		log.Error("[admin] [ws] unknown command", err)
		return errorx.New("unknown command")
	}
	addr := ws.RemoteAddr().String()
	s.watchersLock.Lock()
	s.watchers[addr] = ws
	s.watchersLock.Unlock()

	for {
		mtyp, msg, err = ws.ReadMessage()
		if err != nil {
			return err
		}
		if mtyp != websocket.TextMessage {
			log.Error("[admin] ws read a non TextMessage")
			return errorx.New("not text message")
		}
		if reflectx.BytesToString(msg) == "close" {
			break
		}
	}
	return nil
}
