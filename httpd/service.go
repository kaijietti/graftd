package httpd

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"net"
	"net/http"
	"strings"
)

// Store is the interface that Raft-backend Key-Value Store must implement
type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Del(key string) error
	Join(nodeID, addr string) error
}

// Service provides HTTP service
type Service struct {
	addr   string
	ln     net.Listener
	store  Store
	logger hclog.Logger
}

// New returns an uninitialized HTTP service
func New(addr string, store Store) *Service {
	return &Service{
		addr:   addr,
		store:  store,
		logger: hclog.New(&hclog.LoggerOptions{Name: "graftd-http"}),
	}
}

// Start starts the service
func (s *Service) Start() error {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Error("failed to start", "err", err)
		return err
	}
	s.logger.Info("tcp: successfully start to listen", "addr", s.addr)

	s.ln = ln

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			s.logger.Error("found err when serving", "err", err)
			panic(err)
		}
		s.logger.Info("http: successfully start to serve at", "addr", s.addr)
	}()

	return nil
}

// Close closes the service
func (s *Service) Close() {
	err := s.ln.Close()
	if err != nil {
		s.logger.Error("found err when closing listener", "err", err)
	}
	return
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
	} else if r.URL.Path == "/join" {
		s.handleJoin(w, r)
	} else {
		s.logger.Error("path not found", "path", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		s.logger.Error("found err when unmarshalling join request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(m) != 2 {
		s.logger.Error("join request usage: {'addr': $addr, 'id', $id}")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	remoteAddr, ok := m["addr"]
	if !ok {
		s.logger.Error("join request usage: {'addr': $addr, 'id', $id}")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		s.logger.Error("join request usage: {'addr': $addr, 'id', $id}")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.store.Join(nodeID, remoteAddr); err != nil {
		s.logger.Error("found err when joining node", "node", nodeID, "addr", remoteAddr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.logger.Info("successfully joined node", "node", nodeID, "addr", remoteAddr)
}

func (s *Service) handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	getKey := func() string {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 3 {
			return ""
		}
		return parts[2]
	}

	switch r.Method {
	case http.MethodGet:
		k := getKey()
		if k == "" {
			s.logger.Error("must provide non-empty key")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		v, err := s.store.Get(k)
		if err != nil {
			s.logger.Error("found err when getting key", "key", k, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(map[string]string{k: v})
		if err != nil {
			s.logger.Error("found err when marshalling value", "value", v, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.WriteString(w, string(b))
		if err != nil {
			s.logger.Error("found err when write to resp", "err", err)
		}
		s.logger.Info("OK", "command", fmt.Sprintf("GET <%v>", k))
	case http.MethodPost:
		m := map[string]string{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			s.logger.Error("found err when unmarshalling set request", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for k, v := range m {
			if err := s.store.Set(k, v); err != nil {
				s.logger.Error("found err when unmarshalling set request", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			s.logger.Info("OK", "command", fmt.Sprintf("SET <%v, %v>", k, v))
		}
	case http.MethodDelete:
		k := getKey()
		if k == "" {
			s.logger.Error("must provide non-empty key")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := s.store.Del(k); err != nil {
			s.logger.Error("found err when deleting key", "key", k, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.logger.Info("OK", "command", fmt.Sprintf("DEL <%v>", k))
	default:
		s.logger.Error("path not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}

// Addr returns the address on which the Service is listening
func (s *Service) Addr() net.Addr {
	return s.ln.Addr()
}
