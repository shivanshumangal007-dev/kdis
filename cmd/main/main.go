package main

import (
	"fmt"
	"net"

	"github.com/shivanshumangal007-dev/kdis/internals/helpers"
)

func main() {
	port := ":6379"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("failed to bind port: ", port)
		return
	}

	defer listener.Close()

	fmt.Println("listening on the port: ", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("unable to accept connection")
			return
		}
		// fmt.Println("got one connection:" , conn.LocalAddr())
		go helpers.HandleConnection(conn)
	}
}
