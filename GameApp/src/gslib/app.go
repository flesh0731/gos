package gslib

import (
	"encoding/binary"
	"fmt"
	"goslib/gen_server"
	"goslib/memStore"
	"io"
	"log"
	"net"
	"time"
	"app/register/tables"
	"goslib/broadcast"
	"goslib/logger"
)

func Run() {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("caught panic in main()", x)
		}
	}()

	go SysRoutine()

	time.Sleep(1 * time.Second)

	fmt.Println("Server Started!")

	// Init DB Connections
	memStore.InitDB()
	tables.RegisterTables(memStore.GetSharedDBInstance())

	// Start broadcast server
	gen_server.Start(BROADCAST_SERVER_ID, new(broadcast.Broadcast))

	server_name := "test_player"
	gen_server.Start(server_name, new(Player), server_name)

	// start := time.Now()
	// var times int64 = 1000000
	// var i int64
	// for i = 0; i < times; i++ {
	// 	gen_server.Call(server_name, "hello")
	// }
	// stop := time.Now()
	// fmt.Println("sub: ", float64(stop.Sub(start).Seconds()))
	// fmt.Println("nanoseconds: ", stop.UnixNano()-start.UnixNano())
	// fmt.Println("per seconds: ", float64(times)/stop.Sub(start).Seconds())

	start_tcp_server()
}

func start_tcp_server() {
	l, err := net.Listen("tcp", ":4100")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer func() {
		if x := recover(); x != nil {
			logger.ERR("caught panic in handleClient", x)
		}
	}()

	server_name := conn.RemoteAddr().String()
	gen_server.Start(server_name, new(Player), server_name)

	header := make([]byte, 2)

	for {
		// header
		conn.SetReadDeadline(time.Now().Add(TCP_TIMEOUT * time.Second))
		n, err := io.ReadFull(conn, header)
		if err != nil {
			logger.NOTICE("Connection Closed: ", err)
			break
		}

		// data
		size := binary.BigEndian.Uint16(header)
		data := make([]byte, size)
		n, err = io.ReadFull(conn, data)
		if err != nil {
			logger.WARN("error receiving msg, bytes:", n, "reason:", err)
			break
		}

		gen_server.Cast(server_name, "handleRequest", data, conn)
	}

	gen_server.Cast(server_name, "removeConn")

}