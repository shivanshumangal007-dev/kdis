package raftfsm

import (
	"bufio"
	"bytes"
	"io"

	"github.com/hashicorp/raft"
	"github.com/shivanshumangal007-dev/kdis/internals/helpers"
	"github.com/shivanshumangal007-dev/kdis/internals/store"
)

type KdisFSM struct {
	store *store.InMemoryStore
}

func NewKdisFSM(s *store.InMemoryStore) *KdisFSM {
	return &KdisFSM{store: s}
}

func (f *KdisFSM) Apply(entry *raft.Log) interface{} {
	reader := bufio.NewReader(bytes.NewReader(entry.Data))
	args, err := helpers.ReadCommand(reader) // the exact same parser you've used since Milestone 1
	if err != nil {
		return err
	}
	return helpers.Dispatch(args, f.store, nil)
}
func (f *KdisFSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}
func (f *KdisFSM) Restore(rc io.ReadCloser) error {
	return nil
}
