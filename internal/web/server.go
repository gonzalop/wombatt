package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"wombatt/internal/common"

	"golang.org/x/exp/maps"
)

type page map[string]any

type Server struct {
	root    string
	started bool
	err     error

	pages     map[string]page
	pagesLock sync.RWMutex

	address string
	server  *http.Server
}

func NewServer(address string, root string) *Server {
	p := make(map[string]page)
	if root == "" {
		root = "/"
	} else if root[len(root)-1] != '/' {
		root = root + "/"
	}
	return &Server{
		pages:   p,
		root:    root,
		address: address,
		server:  &http.Server{},
	}
}

func (ls *Server) Start() error {
	ln, err := net.Listen("tcp", ls.address)
	if err != nil {
		return err
	}
	http.Handle(ls.root, ls)
	go func() {
		fmt.Printf("Listening on %v\n", ln.Addr())
		err := ls.server.Serve(ln)
		ls.err = err
		if err != nil {
			slog.Error("error calling TCP Serve", "error", err)
		}
	}()
	ls.started = true
	return nil
}

func (ls *Server) Shutdown(ctx context.Context) error {
	return ls.server.Shutdown(ctx)
}

func (ls *Server) IsRunning() bool {
	return ls.started && ls.err == nil
}

func (ls *Server) Publish(name string, data any) {
	if !ls.IsRunning() {
		return
	}
	if data == nil {
		ls.pagesLock.Lock()
		delete(ls.pages, name)
		ls.pagesLock.Unlock()
		return
	}
	config := make(map[string]interface{})
	f := func(info map[string]string, value any) {
		unit := info["unit"]
		config[info["name"]] = fmt.Sprintf("%v%s", value, unit)
	}
	common.TraverseStruct(data, f)
	name = fmt.Sprintf("%s%s", ls.root, name)
	config["last_updated"] = time.Now().Format(time.RFC3339Nano)
	ls.pagesLock.Lock()
	ls.pages[name] = config
	ls.pagesLock.Unlock()
	slog.Debug("published to web", "url", name)
}

func (ls *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	ls.pagesLock.RLock()
	page, ok := ls.pages[path]
	ls.pagesLock.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	query := r.URL.Query()
	page = filterFields(page, query.Get("fields"))

	if query.Get("format") == "json" {
		j, err := json.Marshal(page)
		if err != nil {
			slog.Error("error formatting json", "path", path, "error", err)
			http.Error(w, "500 server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(j)
		return
	}
	keys := maps.Keys(page)
	sort.Strings(keys)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	for _, k := range keys {
		_, _ = w.Write([]byte(fmt.Sprintf("%s: %v\n", k, page[k])))
	}
	slog.Debug("served from web", "url", path)
}

func filterFields(page map[string]any, fields string) map[string]any {
	if fields == "" {
		return page
	}
	newPage := make(map[string]any)
	for _, k := range strings.Split(fields, ",") {
		if v, ok := page[k]; ok {
			newPage[k] = v
		}
	}
	return newPage
}
