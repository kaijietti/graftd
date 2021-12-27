package store

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const (
	raftTimeout = 10 * time.Second
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Store is a simple raft-backend key-value store
type Store struct {
	RaftDir  string
	RaftBind string
	// used for read kv
	mu sync.Mutex
	kv map[string]string

	// consensus module
	raft *raft.Raft

	logger *log.Logger
}

// New retuens a new store
func New() *Store {
	return &Store{
		kv:     make(map[string]string),
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

// Open opens the store
func (s *Store) Open(bootStrap bool, localID string) error {
	// Setup Raft Configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	// debuging
	config.HeartbeatTimeout = 10000 * time.Millisecond
	config.ElectionTimeout = 10000 * time.Millisecond
	config.CommitTimeout = 500 * time.Millisecond
	config.SnapshotInterval = 120 * time.Second
	config.LeaderLeaseTimeout = 5000 * time.Millisecond

	// setup raft communication
	addr, err := net.ResolveTCPAddr("tcp", s.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftBind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}
	// create snapshot store
	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, 2, os.Stderr)
	if err != nil {
		return err
	}

	// create the log store and stable store
	var logStore raft.LogStore
	var stableStore raft.StableStore

	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(s.RaftDir, "raft.db"))
	if err != nil {
		return err
	}

	stableStore = boltDB
	logStore = boltDB

	ra, err := raft.NewRaft(
		config,
		(*fsm)(s),
		logStore,
		stableStore,
		snapshots,
		transport,
	)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.raft = ra

	if bootStrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

// Store is the interface that Raft-backend Key-Value Store must implement
// type Store interface {
// 	Get(key string) (string, error)
// 	Set(key, value string) error
// 	Del(key string) error
// 	Join(nodeID, addr string) error
// }

// Get returns the value of the given key
func (s *Store) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.kv[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("key not exists")
}

// Set sets the value for the given key
func (s *Store) Set(key, value string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Del deletes the given key
func (s *Store) Del(key string) error {
	if s.raft.State() != raft.Leader {
		return fmt.Errorf("not leader")
	}

	c := &command{
		Op:  "delete",
		Key: key,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(b, raftTimeout)
	return f.Error()
}

// Join joins a node
func (s *Store) Join(nodeID, addr string) error {
	s.logger.Printf("received join request for remote note %s at %s", nodeID, addr)

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	// deal with duplication
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				s.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	s.logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

type fsm Store

func (f *fsm) Apply(l *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		log.Panicf(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}
	switch c.Op {
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDel(c.Key)
	default:
		log.Panicf(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}

	return nil
}

func (f *fsm) applySet(key, value string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.kv[key] = value
	return nil
}

func (f *fsm) applyDel(key string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.kv, key)
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// clone the map
	o := make(map[string]string)
	for k, v := range f.kv {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	f.kv = o
	return nil
}

type fsmSnapshot struct {
	store map[string]string
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// encode data
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		// write data to sink
		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {}
