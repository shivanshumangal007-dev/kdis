package main

import (
	"fmt"
	"net"
	"os"

	"github.com/shivanshumangal007-dev/kdis/internals/helpers"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func main() {
	s := store.NewInMemoryStore()

	if err := helpers.ReaderLineByLine(s); err != nil && !os.IsNotExist(err) {
		fmt.Println("failed to replay access.log:", err)
		return
	}

	port := ":6379"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("failed to bind port: ", port)
		return
	}

	defer listener.Close()

	fmt.Println("listening on the port: ", port)

	go store.ExpiredKeysRemover(s)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("unable to accept connection")
			return
		}
		// fmt.Println("got one connection:" , conn.LocalAddr())
		go helpers.HandleConnection(conn, s)
	}
}
