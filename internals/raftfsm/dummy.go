package raftfsm

import (
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type DummyFSM struct {
	mu   sync.Mutex
	logs []string
}

func (f *DummyFSM) Apply(entry *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.logs = append(f.logs, string(entry.Data))
	fmt.Print("apended entry for entry", entry.Data)
	return nil
}
func (f *DummyFSM) Snapshot() (raft.FSMSnapshot, error) { return nil, nil }
func (f *DummyFSM) Restore(io.ReadCloser) error         { return nil }
