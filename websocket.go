/*
Unlimitedws is a tiny, well-tested library which helps you build a robust Websocket server with a few lines of code.
Specifically, the Websocket server can handle up to several million concurrent connections with only a few GBs of memory.
The library has successfully done that by utilizing async IO, careful tuning OS, zero-upgrade TCP connections, efficient buffer reuse and worker pool in case of DDOS.
*/

package unlimitedws

import (
	"github.com/gobwas/ws"
	"github.com/mailru/easygo/netpoll"
	"log"
	"net"
	"syscall"
	"time"
	"github.com/hxt365/unlimitedws/internal/gopool"
)

// init set rlimit value for maximizing resource limit
func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
}

type Server struct {
	ln              net.Listener
	pool            *gopool.Pool
	poller          netpoll.Poller
	scheduleTimeout time.Duration
	cooldownTime    time.Duration
	ioTimeout       time.Duration
	// OnConnect handles logic when a connection comes up
	OnConnect func(conn net.Conn)
	// OnRead handles logic when a connection sends a message
	OnRead func(conn net.Conn) error
	// OnClose handles logic when a connection gets closed
	OnClose func(conn net.Conn)
}

func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	log.Println("echo server started listening on", ln.Addr().String())

	return &Server{
		ln: ln,
	}, nil
}

// DefaultServer is a echo server with default config
func DefaultServer(addr string) (*Server, error) {
	server, err := NewServer(addr)
	if err != nil {
		return nil, err
	}

	err = server.SetPool(128, 128)
	if err != nil {
		return nil, err
	}

	err = server.SetPoller()
	if err != nil {
		return nil, err
	}

	server.SetScheduleTimeout(time.Millisecond)
	server.SetCooldownTime(5 * time.Millisecond)
	server.SetIOTimeout(100 * time.Millisecond)
	server.OnConnect = func(net.Conn) {}
	server.OnRead = func(net.Conn) error { return nil }
	server.OnClose = func(net.Conn) {}

	return server, err
}

// SetPool set a worker pool for the server
func (s *Server) SetPool(size, queue int) error {
	var err error
	s.pool, err = gopool.NewPool(size, queue)
	return err
}

// SetPoller set a poller for the server
func (s *Server) SetPoller() error {
	var err error
	s.poller, err = netpoll.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

// SetScheduleTimeout set schedule timeout for the server
func (s *Server) SetScheduleTimeout(timeout time.Duration) {
	s.scheduleTimeout = timeout
}

// SetCooldownTime set cooldown time for the server
func (s *Server) SetCooldownTime(time time.Duration) {
	s.cooldownTime = time
}

// SetIOTimeout set io timeout for every readl/write call of connections
// To not set deadline, set timeout to 0
func (s *Server) SetIOTimeout(timeout time.Duration) {
	s.ioTimeout = timeout
}

// Run makes the server ready for accepting request. Here we use asyncIO, with the help of edge-triggered epoll, for the sake of performance.
// In addition, we use worker pool in order to prevent DDOS.
// If the server can not accept incoming request for some reasons (eg: run out of resources), we let the server cool down.
func (s *Server) Run() {
	exit := make(chan struct{})

	acceptDesc := netpoll.Must(netpoll.HandleListener(
		s.ln, netpoll.EventRead|netpoll.EventOneShot,
	))
	accept := make(chan error, 1)

	_ = s.poller.Start(acceptDesc, func(e netpoll.Event) {
		err := s.pool.ScheduleTimeout(s.scheduleTimeout, func() {
			conn, err := s.ln.Accept()
			if err != nil {
				accept <- err
				return
			}
			accept <- nil
			s.handle(conn)
		})

		if err == nil {
			err = <-accept
		}
		if err != nil {
			if err != gopool.ErrScheduleTimeout {
				goto cooldown
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				goto cooldown
			}
		cooldown:
			delay := s.cooldownTime
			time.Sleep(delay)
		}
		_ = s.poller.Resume(acceptDesc)
	})

	<-exit
}

func (s *Server) handle(conn net.Conn) {
	safeConn := conn
	if s.ioTimeout != 0*time.Millisecond {
		safeConn = deadliner{conn, s.ioTimeout}
	}

	// zero-copy upgrade to echo connection
	_, err := ws.Upgrade(safeConn)
	if err != nil {
		_ = conn.Close()
		return
	}
	//u := NewUser(safeConn)
	s.onConnect(conn)

	desc := netpoll.Must(netpoll.HandleRead(conn))
	_ = s.poller.Start(desc, func(event netpoll.Event) {
		if event&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			// When ReadHup or Hup received, this mean that client has closed at least write end of the connection or connections itself
			// So we want to stop receive events about such conn.
			s.onClose(conn, desc)
			return
		}
		// Here we can read msg from the connection, with the support of worker pool.
		s.pool.Schedule(func() {
			s.onRead(safeConn, desc)
		})
	})
}

func (s *Server) onConnect(conn net.Conn) {
	s.OnConnect(conn)
}

func (s *Server) onRead(conn net.Conn, desc *netpoll.Desc) {
	if err := s.OnRead(conn); err != nil {
		s.OnClose(conn)
		_ = s.poller.Stop(desc)
	}
}

func (s *Server) onClose(conn net.Conn, desc *netpoll.Desc) {
	s.OnClose(conn)
	_ = s.poller.Stop(desc)
}

// deadliner is a wapper around net.Conn, which sets read/write timeout for every Read() and Write() call
type deadliner struct {
	net.Conn
	timeout time.Duration
}

func (d deadliner) Read(p []byte) (int, error) {
	if err := d.Conn.SetReadDeadline(time.Now().Add(d.timeout)); err != nil {
		return 0, err
	}
	return d.Conn.Read(p)

}

func (d deadliner) Write(p []byte) (int, error) {
	if err := d.Conn.SetWriteDeadline(time.Now().Add(d.timeout)); err != nil {
		return 0, err
	}
	return d.Conn.Write(p)
}
