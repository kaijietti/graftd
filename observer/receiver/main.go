package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type log struct {
	Offset    int    `json:"offset"`
	Message   string `json:"message"`
	Node      string `json:"node"`
	LogType   string `json:"log_type"`
	Module    string `json:"module"`
	Timestamp int    `json:"unix_timestamp"`
}

func newlog(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	body, _ := ioutil.ReadAll(req.Body)
	log := log{}
	if err := json.Unmarshal(body, &log); err != nil {
		fmt.Println("ERROR")
		return
	}
	fmt.Println(log)
}

func main() {
	http.HandleFunc("/log", newlog)
	http.ListenAndServe(":8090", nil)
}
