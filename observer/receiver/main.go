package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func newlog(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println(string(body))
}

func main() {
	http.HandleFunc("/log", newlog)
	http.ListenAndServe(":8090", nil)
}
