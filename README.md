# Unlimited Websocket

Unlimitedws is a tiny, well-tested library which helps you build a robust Websocket server with a few lines of code. 
Specifically, the Websocket server can handle up to several million concurrent connections with only a few GBs of memory. 
The library has successfully done that by utilizing async IO, careful tuning OS, zero-upgrade TCP connections, efficient buffer reuse and worker pool in case of DDOS.

## Prerequisites
1. Linux OS
2. Some MBs of memory :)

## Installation

```bash
go get github.com/hxt365/unlimitedws
```

## Usage
A robust Echo Websocket system, that handles 100k concurrent connections within only 200 MBs of memory, can be implemented as follows: 
```go
server, _ := unlimitedws.DefaultServer(":8000")

// OnConnect handles logic when a connection comes up
server.OnConnect = func(conn net.Conn) {
    fmt.Println("welcome", nameConn(conn))
}

// OnRead handles logic when a connection sends a message
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

// OnClose handles logic when a connection gets closed
server.OnClose = func(conn net.Conn) {
    fmt.Println("goodbye", nameConn(conn))
}

server.Run()
```
Full example: [Echo](https://github.com/hxt365/unlimitedws/tree/master/examples/echo). 

## Built With
Unlimitedws is built on top of the excellent [Gobwas](https://github.com/gobwas/ws) library. It is highly recommended to read the documentation of that library for advanced usage.

## Planning features
1. Custom Cookie handler, for authentication purpose.
2. Close orphan connections.
3. Websocket fallbacks.
4. Sending missing messages when reconnecting.
5. Horizontal scaling.

## Motivation
1. [Scaling WebSocket in Go and beyond](https://centrifugal.github.io/centrifugo/blog/scaling_websocket/#message-event-stream-benefits) article by Alexander Emelin.
2. [A Million WebSockets and Go](https://www.freecodecamp.org/news/million-websockets-and-go-cc58418460bb/) article by Sergey Kamardin.
3. [Going Infinite, handling 1 millions websockets connections in Go](https://www.youtube.com/watch?v=LI1YTFMi8W4&t=2125s&ab_channel=GopherConIsrael) video by Eran Yanay.


## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://github.com/hxt365/unlimitedws/blob/master/LICENSE.md)