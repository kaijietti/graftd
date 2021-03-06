package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"log"
	"net/http"
	"os"
	"os/signal"

	"graftd/httpd"
	"graftd/store"
	"graftd/utils"
)

// Command line defaults
var (
	DefaultHTTPAddr = "0.0.0.0:11000"
	DefaultRaftAddr = fmt.Sprintf("%s:12000", utils.GetLocalIP())
)

var (
	httpAddr string
	raftAddr string
	joinAddr string
	nodeID   string
	logger   = hclog.New(&hclog.LoggerOptions{
		Name:            "graftd",
		JSONFormat:      true,
		IncludeLocation: true,
	})
)

func init() {
	flag.StringVar(&httpAddr, "haddr", DefaultHTTPAddr, "Set the HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", DefaultRaftAddr, "Set the Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	os.MkdirAll(raftDir, 0700)

	s := store.New()
	s.RaftDir = raftDir
	s.RaftBind = raftAddr
	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := httpd.New(httpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	logger.Info("started successfully!")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	logger.Info("exited successfully")
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		logger.Error("found err when marshalling join request", "err", err)
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		logger.Error("found err when sending join request to node", "node", nodeID, "addr", raftAddr, "err", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
