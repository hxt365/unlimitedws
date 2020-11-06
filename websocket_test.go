package unlimitedws

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	server := newTestDefaultServer(t, ":0")

	go server.Run()

	t.Run("server should accept connection", func(t *testing.T) {
		conn := dialWs(t, server)
		closeWs(t, conn)
	})

	t.Run("bulk connection", func(t *testing.T) {
		numConn := 10000
		for i := 0; i < numConn; i++ {
			conn := dialWs(t, server)
			closeWs(t, conn)
		}
	})
}

func TestServer_SetPool(t *testing.T) {
	server := newTestServer(t, ":0")

	t.Run("test set pool", func(t *testing.T) {
		err := server.SetPool(10, 100)
		assert.NoError(t, err)
		assert.NotEqual(t, nil, server.pool)
	})

	t.Run("test set pool with 0 worker", func(t *testing.T) {
		err := server.SetPool(0, 100)
		assert.Error(t, err)
	})

	t.Run("test set pool with 0 job queue", func(t *testing.T) {
		err := server.SetPool(10, 0)
		assert.Error(t, err)
	})
}

func TestServer_SetPoller(t *testing.T) {
	server := newTestServer(t, ":0")

	err := server.SetPoller()
	assert.NoError(t, err)
	assert.NotEqual(t, nil, server.poller)
}

func TestServer_SetScheduleTimeout(t *testing.T) {
	server := newTestServer(t, ":0")

	timeout := time.Millisecond
	server.SetScheduleTimeout(timeout)
	assert.Equal(t, timeout, server.scheduleTimeout)
}

func TestServer_SetCooldownTime(t *testing.T) {
	server := newTestServer(t, ":0")

	timeout := 5 * time.Millisecond
	server.SetCooldownTime(timeout)
	assert.Equal(t, timeout, server.cooldownTime)
}

func TestServer_SetIOTimeout(t *testing.T) {
	server := newTestServer(t, ":0")

	timeout := 100 * time.Millisecond
	server.SetIOTimeout(timeout)
	assert.Equal(t, timeout, server.ioTimeout)
}

func TestServer_Hook(t *testing.T) {
	server := newTestDefaultServer(t, ":0")
	go server.Run()

	_connect := 0
	_read := 0
	_close := 0
	server.OnConnect = func(net.Conn) { _connect = 1 }
	server.OnRead = func(net.Conn) error {
		_read = 1
		return nil
	}
	server.OnClose = func(net.Conn) { _close = 1 }

	conn := dialWs(t, server)
	err := conn.WriteMessage(websocket.TextMessage, []byte("Hello world"))
	assert.NoError(t, err)
	closeWs(t, conn)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, _connect)
	assert.Equal(t, 1, _read)
	assert.Equal(t, 1, _close)
}

func TestServer_NewDefaultServer(t *testing.T) {
	server, err := DefaultServer(":0")
	assert.NoError(t, err)
	assert.NotEqual(t, nil, server.pool)
	assert.NotEqual(t, nil, server.poller)
	assert.Equal(t, time.Millisecond, server.scheduleTimeout)
	assert.Equal(t, 5*time.Millisecond, server.cooldownTime)
	assert.Equal(t, 100*time.Millisecond, server.ioTimeout)
}

func newTestServer(t testing.TB, addr string) *Server {
	server, err := NewServer(addr)
	assert.NoError(t, err)
	return server
}

func newTestDefaultServer(t testing.TB, addr string) *Server {
	server, err := DefaultServer(addr)
	assert.NoError(t, err)
	return server
}

func TestServer_Echo(t *testing.T) {
	server := newTestDefaultServer(t, ":0")
	server.OnRead = func(conn net.Conn) error {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			return err
		}
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			return err
		}
		return nil
	}
	go server.Run()

	conn := dialWs(t, server)
	err := conn.WriteMessage(websocket.TextMessage, []byte("Hello world"))
	assert.NoError(t, err)
	closeWs(t, conn)

	var msg []byte
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		msg = append(msg, message...)
	}
}

func dialWs(t testing.TB, server *Server) *websocket.Conn {
	wsURL := "ws://" + server.ln.Addr().String()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		assert.NoError(t, err)
	}
	return conn
}

func closeWs(t testing.TB, conn *websocket.Conn) {
	assert.NoError(t, conn.Close())
}
