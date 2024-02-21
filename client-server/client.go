package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func UNUSED(x ...interface{}) {}

const (
	ADDR_SERVER = "62.217.185.17:8080"
	END_BYTES   = "\000\001\002\003\004\005"
)

func main() {
	conn, err := net.Dial("tcp", ADDR_SERVER)

	if err != nil {
		panic("Couldn't connect")
	}
	defer conn.Close()
	go ClientOutput(conn)
	ClientInput(conn)
}

func ClientInput(conn net.Conn) {
	for {
		conn.Write([]byte(InputString() + END_BYTES))
	}
}

func ClientOutput(conn net.Conn) {
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
			// message += string(buffer[:length])
		}
		fmt.Println(message)
	}

}

func InputString() string {
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	msg = strings.Replace(msg, "\n", "", -1)

	return msg
}
