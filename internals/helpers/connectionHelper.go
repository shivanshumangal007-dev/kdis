package helpers

import (
	"bufio"
	"net"

	"github.com/shivanshumangal007-dev/kdis/internals/resp"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func HandleConnection(conn net.Conn, s *store.InMemoryStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// buf := make([]byte, 1024)
	for {
		args, err := readCommand(reader)
		if err != nil {
			return
		}

		ans := dispatch(args, s)
		conn.Write([]byte(ans))
	}
}

func dispatch(args []string, s *store.InMemoryStore) string {
	if len(args) == 0 {
		return "-ERR empty commands\r\n"
	}

	return resp.RespReplyEncoder(args, s)
}
