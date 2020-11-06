package main

import (
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"github.com/hxt365/unlimitedws"
)

var count int64

func main() {
	openPprof()

	server, _ := unlimitedws.DefaultServer(":8000")

	server.OnConnect = func(conn net.Conn) {
		n := atomic.AddInt64(&count, 1)
		if n%1000 == 0 {
			fmt.Println("Total number of concurrent connections:", n)
		}
	}

	server.OnRead = func(conn net.Conn) error {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			return err
		}
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			return err
		}
		fmt.Println(string(msg))
		return nil
	}

	server.OnClose = func(conn net.Conn) {
		n := atomic.AddInt64(&count, -1)
		if n%1000 == 0 {
			fmt.Println("Total number of concurrent connections:", n)
		}
	}

	server.Run()
}

func openPprof() {
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("Pprof failed: %v", err)
		}
	}()
}
