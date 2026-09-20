package helpers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// buf := make([]byte, 1024)
	for {
		args, err := readCommand(reader)
		if err != nil {
			return
		}

		ans := dispatch(args)
		conn.Write([]byte(ans))
	}
}

func dispatch(args []string) string {
	if len(args) == 0 {
		return "-ERR empty commands"
	}
	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "PING":
		return "+PONG\r\n"
	case "ECHO":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'echo' command\r\n"
		}
		return fmt.Sprintf("$%d\r\n%s\r\n", len(args[1]), args[1])
	default:
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd)
	}
}
