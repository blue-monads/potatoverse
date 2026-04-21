package binds

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type BindingServer struct {
	app       xtypes.App
	installId int64
	spaceId   int64
	port      int
	server    *http.Server
	contexts  *sync.Map
}

func NewBindingServer(app xtypes.App, installId, spaceId int64, port int) *BindingServer {
	return &BindingServer{
		app:       app,
		installId: installId,
		spaceId:   spaceId,
		port:      port,
		contexts:  &sync.Map{},
	}
}

func (s *BindingServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/call", s.handleCall)

	s.server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			qq.Println("@binding_server_error", err)
		}
	}()

	return nil
}

func (s *BindingServer) Stop() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *BindingServer) RegisterContext(id string, event *xtypes.HttpEvent) {
	s.contexts.Store(id, event)
}

func (s *BindingServer) UnregisterContext(id string) {
	s.contexts.Delete(id)
}

func (s *BindingServer) handleCall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Module string `json:"module"`
		Method string `json:"method"`
		Params []any  `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestId := r.Header.Get("X-Potato-Request-ID")
	var ctx *xtypes.HttpEvent
	if requestId != "" {
		if val, ok := s.contexts.Load(requestId); ok {
			ctx = val.(*xtypes.HttpEvent)
		}
	}

	var result any
	var err error

	switch req.Module {
	case "db":
		result, err = s.handleDB(req.Method, req.Params)
	case "kv":
		result, err = s.handleKV(req.Method, req.Params)
	case "signer":
		result, err = s.handleSigner(req.Method, req.Params, ctx)
	default:
		err = fmt.Errorf("unknown module: %s", req.Module)
	}

	resp := map[string]any{
		"result": result,
	}
	if err != nil {
		resp["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *BindingServer) GetPort() int {
	return s.port
}
