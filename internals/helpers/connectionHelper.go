package helpers

import (
	"bufio"
	"net"

	"github.com/shivanshumangal007-dev/kdis/internals/resp"
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

	return resp.RespReplyDecoder(args)
}
