package server

import (
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/recallsong/go-utils/net/echox"
	"github.com/recallsong/httpc"
	log "github.com/sirupsen/logrus"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (s *Server) initHttpRoutes(svr *echox.EchoServer) {
	svr.GET("/hello", s.Hello)
	svr.GET("/wsconn", s.wsConn)
}

func (s *Server) Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello !")
}

func (s *Server) wsConn(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	s.onConn(&wsConn{ws, nil})
	return nil
}

type Conn struct {
	Conn    net.Conn
	StartOn time.Time
	Pkg     *connectPkg
}

func (s *Server) onConn(conn net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Error(err)
			conn.Close()
		}
	}()
	if cc, ok := conn.(*net.TCPConn); ok {
		cc.SetNoDelay(true)
	}
	keepalive := s.cfg.KeepAlive * 2
	reader, writer := conn, conn
	buffer := make([]byte, 1024, 1024)
	addr := conn.RemoteAddr().String()

	// 连接握手
	connpkg := &connectPkg{}
	conn.SetReadDeadline(time.Now().Add(keepalive))
	err := readConnPackage(reader, buffer, connpkg)
	if err != nil {
		log.Errorf("[conn] %s read connectPkg error: %s", addr, err)
		closeConn(conn)
		return
	}
	connackpkg := &connAckPkg{Code: 200, Keepalive: int64(s.cfg.KeepAlive / time.Second)}
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err = writeConnAckPackage(writer, buffer, connackpkg)
	if err != nil {
		log.Errorf("[conn] %s write connAckPkg error: %s", addr, err)
		closeConn(conn)
		return
	}
	err = s.checkConn(addr, connpkg)
	if err != nil {
		log.Errorf("[conn] fail to check , addr:%s , topic:%s , tkn:%s , dmn:%s", addr, connpkg.Topic, connpkg.Token, connpkg.Domain)
		closeConn(conn)
		return
	}
	c := &Conn{Conn: conn, StartOn: time.Now(), Pkg: connpkg}
	connpkg, connackpkg = nil, nil
	s.onConnEvent(addr, c)
	atomic.AddInt32(&s.connNum, 1)
	defer atomic.AddInt32(&s.connNum, -1)

	// 开始读取心跳包
	for {
		conn.SetReadDeadline(time.Now().Add(keepalive))
		err = readPingPackage(reader, buffer)
		if err != nil {
			closeConn(conn)
			s.onCloseEvent(addr, c)
			return
		}
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err = writePongPackage(writer, buffer)
		if err != nil {
			closeConn(conn)
			s.onCloseEvent(addr, c)
			return
		}
	}
}

func closeConn(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Error("[conn] close connect error: ", err)
	}
}

func (s *Server) onConnEvent(addr string, conn *Conn) {
	s.connsLock.Lock()
	s.conns[addr] = conn
	s.connsLock.Unlock()
	log.Infof("[conn] %s connected, topic:%s, tkn:%s, dmn:%s", addr, conn.Pkg.Topic, conn.Pkg.Token, conn.Pkg.Domain)
	s.notifyToAdminPage(addr, conn, "online")
	if s.store != nil {
		s.pushOnlineEvent(addr, conn)
	}
}

func (s *Server) onCloseEvent(addr string, conn *Conn) {
	s.connsLock.Lock()
	delete(s.conns, addr)
	s.connsLock.Unlock()
	log.Infof("[conn] %s disconnected, topic:%s, tkn:%s, dmn:%s", addr, conn.Pkg.Topic, conn.Pkg.Token, conn.Pkg.Domain)
	s.notifyToAdminPage(addr, conn, "offline")
	if s.store == nil {
		s.pushOfflineEvent(addr, conn)
	}
}

var errConnCheckFail = errors.New("connect check failed")

func (s *Server) checkConn(addr string, cpkg *connectPkg) error {
	if len(s.cfg.ConnCheckUrl) <= 0 {
		return nil
	}
	var status int
	err := httpc.New(s.cfg.ConnCheckUrl).SuccessStatus(-500).
		Body(cpkg, httpc.TypeApplicationJson).Post(&status)
	if err != nil {
		log.Error("[conn] check request error: ", err)
		return err
	}
	if status != 200 {
		return errConnCheckFail
	}
	return nil
}
