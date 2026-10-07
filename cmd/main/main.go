package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/shivanshumangal007-dev/kdis/internals/helpers"
	"github.com/shivanshumangal007-dev/kdis/internals/pubsub"
	"github.com/shivanshumangal007-dev/kdis/internals/raftfsm"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

func main() {
	s := store.NewInMemoryStore()
	ps := pubsub.NewPubsubStore()
	// flags
	nodeID := flag.String("id", "node1", "this node's ID")
	raftAddr := flag.String("raftaddr", "127.0.0.1:7000", "this node's raft address")
	dataDir := flag.String("datadir", "raft-data-node1", "raft's own data dir")
	respAddr := flag.String("respaddr", ":6379", "where redis-cli connects")
	flag.Parse()

	// raft setup
	os.MkdirAll(*dataDir, 0755)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(*nodeID)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", *raftAddr)
	transport, err := raft.NewTCPTransport(*raftAddr, tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		fmt.Println("transport error:", err)
		return
	}
	logStore, _ := raftboltdb.NewBoltStore(*dataDir + "/raft-log.bolt")
	stableStore, _ := raftboltdb.NewBoltStore(*dataDir + "/raft-stable.bolt")
	snapshotStore, _ := raft.NewFileSnapshotStore(*dataDir, 1, os.Stderr)

	fsm := raftfsm.NewKdisFSM(s) // ← the one real change from your test code

	raftNode, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		fmt.Println("raft init error:", err)
		return
	}

	hasState, _ := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if !hasState {
		raftNode.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: "node1", Address: "127.0.0.1:7000"},
				{ID: "node2", Address: "127.0.0.1:7001"},
				{ID: "node3", Address: "127.0.0.1:7002"},
			},
		})
	}
	port := *respAddr
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
		go helpers.HandleConnection(conn, s, ps)
	}
}
