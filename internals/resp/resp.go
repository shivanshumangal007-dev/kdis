package resp

import (
	"fmt"
	"strings"

	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func RespReplyEncoder(args []string, s *store.InMemoryStore) string {
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
