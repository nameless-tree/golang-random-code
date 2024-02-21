package main

import (
	"log"
	"net"
	"strings"
)

const (
	// ADDR_SERVER = ":8080"
	END_BYTES = "\000\001\002\003\004\005"
	PORT      = ":8080"
)

var (
	Connections = make(map[net.Conn]bool)
)

func main() {

	// init connection
	listen, err := net.Listen("tcp", PORT)

	if err != nil {
		panic("Server error")
	}

    log.Println("listening...", listen)

	defer listen.Close()

	for {
		// listening
		conn, err := listen.Accept()
		if err != nil {
			break
		}

		go handeConnect(conn)
	}
}

func handeConnect(conn net.Conn) {
	Connections[conn] = true

	var (
		buffer  = make([]byte, 512)
		message string
	)

close:
	for {
		message = ""

		for {
			length, err := conn.Read(buffer)

			if err != nil {
				break close
			}

			message += string(buffer[:length])

			if strings.HasSuffix(message, END_BYTES) {
				message = strings.TrimSuffix(message, END_BYTES)
				break
			}
		}

		log.Println(message)

		for c := range Connections {
			// do not send to sender
			if c == conn {
				continue
			}
			c.Write([]byte(strings.ToUpper(message) + END_BYTES))

		}
	}
	delete(Connections, conn)
}
