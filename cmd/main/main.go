package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
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

func readCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")

	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected array, got: %q", line)
	}
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}

	args := make([]string, count)
	for i := 0; i < count; i++ {
		bulkheader, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		bulkheader = strings.TrimRight(bulkheader, "\r\n")
		length, err := strconv.Atoi(bulkheader[1:])
		if err != nil {
			return nil, err
		}

		buf := make([]byte, length+2)
		_, err = readFull(reader, buf)
		if err != nil {
			return nil, err
		}
		args[i] = string(buf[:length])

	}

	return args, nil
}

func readFull(reader *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := reader.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
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
