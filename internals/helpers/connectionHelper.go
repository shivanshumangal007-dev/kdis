package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"github.com/shivanshumangal007-dev/kdis/internals/pubsub"
	"github.com/shivanshumangal007-dev/kdis/internals/resp"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func HandleConnection(conn net.Conn, s *store.InMemoryStore, ps *pubsub.PubsubStore, raftNode *raft.Raft) {
	defer conn.Close()
	var subChan []chan string
	defer func() {
		for _, ch := range subChan {
			if ch != nil {
				ps.QuitSubsciber(ch)
			}
		}
	}()
	reader := bufio.NewReader(conn)
	// buf := make([]byte, 1024)
	for {
		args, err := ReadCommand(reader)
		if err != nil {
			return
		}

		// if shouldPersist(args) {
		// 	if err := Writter(args); err != nil {
		// 		return
		// 	}
		// }
		var ans string
		cmd := strings.ToUpper(args[0])
		switch {
		case cmd == "SUBSCRIBE":
			subChannel, err := dispatchSubs(args, ps, conn)
			if err != nil {
				conn.Write([]byte(err.Error()))
			} else {
				subChan = append(subChan, subChannel)
			}
		case isReplicatedCommand(args):
			if raftNode.State() != raft.Leader {
				_, leaderID := raftNode.LeaderWithID()
				conn.Write([]byte(fmt.Sprintf("-ERR not leader, try %s\r\n", leaderID)))
				continue
			}
			encoded := EncodeRESPCommand(args)
			future := raftNode.Apply(encoded, 5*time.Second)
			if err := future.Error(); err != nil {
				conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err)))
				continue
			}
			result := future.Response()
			ans, ok := result.(string)
			if !ok {
				conn.Write([]byte(fmt.Sprintf("-ERR unexpected result: %v\r\n", result)))
				continue
			}
			conn.Write([]byte(ans))
		default:
			ans = Dispatch(args, s, ps)
			conn.Write([]byte(ans))
		}
	}
}

func isReplicatedCommand(args []string) bool {
	// return false               //uncomment this line to turn off writter to the file
	switch strings.ToUpper(args[0]) {
	case "SET", "DEL", "EXPIRE", "LPUSH", "RPUSH", "HSET", "SADD":
		return true
	default:
		return false
	}
}

func EncodeRESPCommand(args []string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&buf, "$%d\r\n%s\r\n", len(a), a)
	}
	return buf.Bytes()
}

func Dispatch(args []string, s *store.InMemoryStore, ps *pubsub.PubsubStore) string {
	if len(args) == 0 {
		return "-ERR empty commands\r\n"
	}

	return resp.RespReplyEncoder(args, s, ps)
}
func dispatchSubs(args []string, ps *pubsub.PubsubStore, conn net.Conn) (chan string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("-ERR empty commands\r\n")
	}
	cmd := strings.ToUpper(args[0])
	if cmd != "SUBSCRIBE" {
		return nil, fmt.Errorf("-WRONG Dispatch funcion\r\n")
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("-ERR wrong number of arguments for 'SUBSCRIBE' command\r\n")
	}
	subchannelName := args[1]
	newSubs := ps.NewSubsciber(subchannelName)
	go func() {
		for msg := range newSubs {
			reply := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(subchannelName), subchannelName, len(msg), msg)
			conn.Write([]byte(reply))
		}
	}()

	conn.Write(fmt.Appendf([]byte{}, "*3\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n:1\r\n",
		len(subchannelName), subchannelName))

	return newSubs, nil
}
