package server

import (
	"errors"
	"io"
	"time"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	*websocket.Conn
	reader io.Reader
}

func (ws *wsConn) SetDeadline(time.Time) error {
	return errors.New("SetDeadline not support")
}

// 从websocket中读取二进制数据，从reader中读取未读完的数据
func (ws *wsConn) Read(b []byte) (int, error) {
	if ws.reader != nil {
		n, err := ws.reader.Read(b)
		if err == io.EOF && n == 0 {
			ws.reader = nil
		} else {
			return n, err
		}
	}
	t, r, err := ws.NextReader()
	if err != nil {
		return 0, err
	} else if t != websocket.BinaryMessage {
		return 0, errors.New("not binary message")
	} else {
		ws.reader = r
		return r.Read(b)
	}
}

func (ws *wsConn) Write(b []byte) (int, error) {
	w, err := ws.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	w.Close()
	return n, err
}
