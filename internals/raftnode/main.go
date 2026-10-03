package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/shivanshumangal007-dev/kdis/internals/raftfsm"
)

func main() {
	nodeID := flag.String("id", "node1", "this node's ID")
	raftAddr := flag.String("raftaddr", "127.0.0.1:7000", "this node's raft address")
	dataDir := flag.String("datadir", "raft-data-node1", "where this node stores its data")
	flag.Parse()

	os.MkdirAll(*dataDir, 0755)

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(*nodeID)

	addr := *raftAddr
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	transport, err := raft.NewTCPTransport(addr, tcpAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		fmt.Println("transport error:", err)
		return
	}

	logStore, _ := raftboltdb.NewBoltStore(*dataDir + "/raft-log.bolt")
	stableStore, _ := raftboltdb.NewBoltStore(*dataDir + "/raft-stable.bolt")
	snapshotStore, _ := raft.NewFileSnapshotStore(*dataDir, 1, os.Stderr)

	fsm := &raftfsm.DummyFSM{}
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		fmt.Println("raft init error:", err)
		return
	}
	hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		fmt.Println("error checking state:", err)
		return
	}

	if !hasState {
		r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{ID: "node1", Address: "127.0.0.1:7000"},
				{ID: "node2", Address: "127.0.0.1:7001"},
				{ID: "node3", Address: "127.0.0.1:7002"},
			},
		})
	}

	// give it a moment to elect itself leader
	time.Sleep(2 * time.Second)

	if r.State() == raft.Leader {
		fmt.Println("I am the leader")
		future := r.Apply([]byte("hello world"), 5*time.Second)
		if err := future.Error(); err != nil {
			fmt.Println("apply error:", err)
		} else {
			fmt.Println("write committed")
		}
	}

	select {} // block forever, keep the node alive
}
