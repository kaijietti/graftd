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
	"strings"
)

// Easier to get running with CORS. Thanks for help @Vindexus and @erkie
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

// logs from logstash
type lstLog struct {
	Node   string `json:"node"`
	Offset int    `json:"offset"`
	//Timestamp string `json:"unix_timestamp"`
}

type logInfo struct {
	lstLog
	Message string `json:"message"`
}

// logstashLog.message is hc-log json
type moduleLog struct {
	lstLog
	Caller        string                 `json:"caller"`
	TimestampNano string                 `json:"unix_timestamp"`
	Level         string                 `json:"level"`
	Module        string                 `json:"module"`
	Message       string                 `json:"message"`
	Extend        map[string]interface{} `json:"extend"`
}

func (m *moduleLog) ConvertFrom(lf *logInfo) {

	extend := make(map[string]interface{})

	unescapedStr := strings.ReplaceAll(lf.Message, "\\\"", "\"")
	if err := json.Unmarshal([]byte(unescapedStr), &extend); err != nil {
		fmt.Println(err)
		return
	}

	// basic info
	m.Node = lf.Node
	m.Offset = lf.Offset
	//m.Timestamp = lf.Timestamp

	// hc-log info
	m.Caller = extend["@caller"].(string)
	m.Level = extend["@level"].(string)
	m.Module = extend["@module"].(string)
	m.Message = extend["@message"].(string)
	m.TimestampNano = extend["@timestamp"].(string)

	// delete some keys
	delete(extend, "@caller")
	delete(extend, "@level")
	delete(extend, "@module")
	delete(extend, "@message")
	delete(extend, "@timestamp")

	m.Extend = extend
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

	ch := make(chan *moduleLog, 10000)

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

		fmt.Printf("lst-log: %v\n", string(body))

		log := logInfo{}
		if err := json.Unmarshal(body, &log); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("raw-log: %+v\n", log)

		hlog := &moduleLog{}
		hlog.ConvertFrom(&log)
		//hlog.Timestamp = hlog.TimestampNano

		fmt.Printf("module-log: %+v\n\n", *hlog)

		ch <- hlog
	})

	log.Println("Serving at localhost:8090...")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
