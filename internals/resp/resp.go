package resp

import (
	"fmt"
	"strconv"
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

	case "SET":
		if len(args) != 3 { // SET key value
			return "-ERR wrong number of arguments for 'SET' command\r\n"
		}
		s.Set(args[1], args[2])
		return "+OK\r\n"

	case "GET":
		if len(args) != 2 { // SET key value
			return "-ERR wrong number of arguments for 'GET' command\r\n"
		}
		val, found, err := s.Get(args[1])
		if err != nil{
			return fmt.Sprintf("-%s\r\n", err)
		}
		if found == true {
			return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		}
		return "$-1\r\n"

	case "DEL":
		if len(args) != 2 { // SET key value
			return "-ERR wrong number of arguments for 'DEL' command\r\n"
		}
		found := s.Del(args[1])
		if found == true {
			return ":1\r\n"
		}
		return ":0\r\n"
	case "EXISTS":
		if len(args) != 2 { // SET key value
			return "-ERR wrong number of arguments for 'EXISTS' command\r\n"
		}
		found := s.Exists(args[1])
		if found == true {
			return ":1\r\n"
		}
		return ":0\r\n"
	case "EXPIRE":
		if len(args) != 3 { // SET key value
			return "-ERR wrong number of arguments for 'EXPIRES' command\r\n"
		}
		second, err := strconv.Atoi(args[2])
		if err != nil {
			return "-ERR value is not an integer or out of range\r\n"
		}
		ans := s.Expire(args[1], second)
		if ans {
			return ":1\r\n"
		}
		return ":0\r\n"

	case "TTL":
		if len(args) != 2 { // SET key value
			return "-ERR wrong number of arguments for 'TTL' command\r\n"
		}
		remain := s.Ttl(args[1])
		return fmt.Sprintf(":%d\r\n", remain)
	
	case "LPUSH":
		if len(args) < 3{ // LPUSH key value values[]
			return "-ERR wrong number of arguments for 'LPUSH' command\r\n"
		}

		len, err := s.Lpush(args[1], args[2:]...)
		if err != nil{
			return fmt.Sprintf("-%s\r\n", err)
		}

		return fmt.Sprintf(":%d\r\n", len)
	case "RPUSH":
		if len(args) < 3{ // LPUSH key value values[]
			return "-ERR wrong number of arguments for 'RPUSH' command\r\n"
		}

		len, err := s.Rpush(args[1], args[2:]...)
		if err != nil{
			return fmt.Sprintf("-%s\r\n", err)
		}

		return fmt.Sprintf(":%d\r\n", len)

	default:
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd)
	}
}
