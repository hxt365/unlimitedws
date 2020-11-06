package main

import (
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"net"
	"github.com/hxt365/unlimitedws"
)

func main() {
	server, _ := unlimitedws.DefaultServer(":8000")
	server.OnConnect = func(conn net.Conn) {
		fmt.Println("welcome", nameConn(conn))
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
		fmt.Println("goodbye", nameConn(conn))
	}
	server.Run()
}

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}