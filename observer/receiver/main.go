package main

import (
	"encoding/json"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"io/ioutil"
	"log"
	"net/http"
)

// Easier to get running with CORS. Thanks for help @Vindexus and @erkie
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

type RaftState struct {
	// The current term, cache of StableStore
	currentTerm uint64

	// Highest committed log entry
	commitIndex uint64

	// Last applied log to the FSM
	lastApplied uint64

	// Cache the latest snapshot index/term
	lastSnapshotIndex uint64
	lastSnapshotTerm  uint64

	// Cache the latest log from LogStore
	lastLogIndex uint64
	lastLogTerm  uint64

	// The current state
	state string
}

type logInfo struct {
	Offset    int       `json:"offset"`
	Message   string    `json:"message"`
	Node      string    `json:"node"`
	LogType   string    `json:"log_type"`
	Module    string    `json:"module"`
	Timestamp int       `json:"unix_timestamp"`
	RaftState RaftState `json:"-"`
}

func main() {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	ch := make(chan *logInfo, 10)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		go func() {
			for log := range ch {
				s.Emit("append log", log)
			}
		}()
		return nil
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	// socket.io
	http.Handle("/socket.io/", server)
	// static files
	http.Handle("/", http.FileServer(http.Dir("./assets")))

	http.HandleFunc("/log", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := ioutil.ReadAll(req.Body)
		log := logInfo{}
		if err := json.Unmarshal(body, &log); err != nil {
			fmt.Println("ERROR")
			return
		}
		if err := json.Unmarshal([]byte(log.Message), &log.RaftState); err != nil {

		}
		ch <- &log
		fmt.Println(log)
	})

	log.Println("Serving at localhost:8090...")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
