package helpers

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/shivanshumangal007-dev/kdis/internals/pubsub"
	"github.com/shivanshumangal007-dev/kdis/internals/resp"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

type command struct{
	args []string
	replyChl chan string
}

func HandleConnection(conn net.Conn, s *store.InMemoryStore, ps *pubsub.PubsubStore) {
	defer conn.Close()
	var subChan []chan string
	defer func() {
		for _, ch := range subChan{
			if ch != nil {
				ps.QuitSubsciber(ch)
			}
		}
	}()
	reader := bufio.NewReader(conn)
	// buf := make([]byte, 1024)
	for {
		args, err := readCommand(reader)
		if err != nil {
			return
		}

		if shouldPersist(args) {
			if err := Writter(args); err != nil {
				return
			}
		}
		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "SUBSCRIBE":
			subChannel, err := dispatchSubs(args, ps, conn)
			if err != nil {
				conn.Write([]byte(err.Error()))
			}else {
				subChan = append(subChan, subChannel)
			}
		default:
			// ans := dispatch(args, s, ps)

			replyChl := make(chan string)
			commandQueue <- command{args: args, replyChl: replyChl}
			ans :=  <- replyChl
			
			conn.Write([]byte(ans))
		}
	}
}

func shouldPersist(args []string) bool {
	// return false               //uncomment this line to turn off writter to the file
	switch strings.ToUpper(args[0]) {
	case "SET", "DEL", "EXPIRE", "LPUSH", "RPUSH", "HSET", "SADD":
		return true
	default:
		return false
	}
}

func dispatch(args []string, s *store.InMemoryStore, ps *pubsub.PubsubStore) string {
	if len(args) == 0 {
		return "-ERR empty commands\r\n"
	}

	return resp.RespReplyEncoder(args, s, ps)
}
func dispatchSubs(args []string, ps *pubsub.PubsubStore, conn net.Conn) (chan string ,error) {
	if len(args) == 0 {
		return nil ,fmt.Errorf("-ERR empty commands\r\n")
	}
	cmd := strings.ToUpper(args[0])
	if cmd != "SUBSCRIBE" {
		return nil,fmt.Errorf("-WRONG dispatch funcion\r\n")
	}
	if len(args) != 2 {
		return nil,fmt.Errorf("-ERR wrong number of arguments for 'SUBSCRIBE' command\r\n")
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

var commandQueue = make(chan command, 1000)

func Executer(s *store.InMemoryStore , ps *pubsub.PubsubStore) {
	fmt.Print("executer started for v2 check")
	for cmd := range commandQueue {
		ans := dispatch(cmd.args, s, ps)
		rep := cmd.replyChl
		rep <- ans
	}
}