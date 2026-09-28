package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func respSetter(args []string) string {
	var str strings.Builder
	fmt.Fprintf(&str, "*%d\r\n", len(args))
	for _, val := range args {
		fmt.Fprintf(&str, "$%d\r\n%s\r\n", len(val), val)
	}
	return str.String()
}
func goConnection(id int, wg *sync.WaitGroup) {
	fmt.Printf("strted worker with id: %d\n", id)
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal("connection refused")
	}
	defer wg.Done()
	defer conn.Close()
	cnt := 1000
	reader := bufio.NewReader(conn)
	for i := 0; i < cnt; i++ {

		key := fmt.Sprintf("key-%d-%d", id, i)
		value := fmt.Sprintf("val-%d-%d", id, i)
		resp := respSetter([]string{"SET", key, value})
		_, err := conn.Write([]byte(resp))
		if err != nil {
			log.Printf("worker %d: write failed: %v", id, err)
			return
		}
		reply, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("worker %d: read failed: %v", id, err)
			return
		}
		if reply != "+OK\r\n" {
			log.Printf(
				"worker %d: unexpected response: %q",
				id,
				reply,
			)
			return
		}
	}
}
func main() {
	start := time.Now()
	var wg sync.WaitGroup
	workers := 30
	commandsPerWorker := 1000

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go goConnection(i, &wg)
	}
	wg.Wait()

	elapsed := time.Since(start)
	total := workers * commandsPerWorker

	fmt.Printf("Total commands: %d\n", total)
	fmt.Printf("Elapsed: %v\n", elapsed)
	fmt.Printf("Throughput: %.2f commands/sec\n",
		float64(total)/elapsed.Seconds(),
	)

}
