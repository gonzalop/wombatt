package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"wombatt/internal/common"
)

//go:embed static
var staticFiles embed.FS

type page map[string]any

type Server struct {
	root    string
	started bool
	err     error

	pages     map[string]page
	rawPages  map[string]map[string]any
	pagesLock sync.RWMutex

	address string
	server  *http.Server
}

func NewServer(address string, root string) *Server {
	p := make(map[string]page)
	rp := make(map[string]map[string]any)
	if root == "" {
		root = "/"
	} else if root[len(root)-1] != '/' {
		root = root + "/"
	}
	return &Server{
		pages:    p,
		rawPages: rp,
		root:     root,
		address:  address,
		server:   &http.Server{},
	}
}

func (ls *Server) Start() error {
	ln, err := net.Listen("tcp", ls.address)
	if err != nil {
		return err
	}
	// Always handle root to serve static files and dashboard
	http.Handle("/", ls)
	http.HandleFunc("/metrics", ls.ServeMetrics)
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
		delete(ls.rawPages, name)
		ls.pagesLock.Unlock()
		return
	}
	config := make(map[string]any)
	rawConfig := make(map[string]any)
	f := func(info map[string]string, value any) {
		unit := info["unit"]
		config[info["name"]] = fmt.Sprintf("%v%s", value, unit)
		rawConfig[info["name"]] = value
	}
	common.TraverseStruct(data, f)
	name = fmt.Sprintf("%s%s", ls.root, name)
	config["last_updated"] = time.Now().Format(time.RFC3339Nano)
	ls.pagesLock.Lock()
	ls.pages[name] = config
	ls.rawPages[name] = rawConfig
	ls.pagesLock.Unlock()
	slog.Debug("published to web", "url", name)
}

func (ls *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	ls.pagesLock.RLock()
	page, ok := ls.pages[path]
	ls.pagesLock.RUnlock()
	if !ok {
		// Serve static files
		fsys, err := fs.Sub(staticFiles, "static")
		if err != nil {
			slog.Error("error in fs.Sub", "error", err)
			http.Error(w, "500 server error", http.StatusInternalServerError)
			return
		}
		http.FileServer(http.FS(fsys)).ServeHTTP(w, r)
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
	keys := slices.Sorted(maps.Keys(page))
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	for _, k := range keys {
		_, _ = w.Write(fmt.Appendf(nil, "%s: %v\n", k, page[k]))
	}
	slog.Debug("served from web", "url", path)
}

func (ls *Server) ServeMetrics(w http.ResponseWriter, r *http.Request) {
	ls.pagesLock.RLock()
	defer ls.pagesLock.RUnlock()

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)

	for path, page := range ls.rawPages {
		// Use full path as source so dashboard can find it
		source := path
		for key, value := range page {
			val := reflect.ValueOf(value)
			var floatVal float64
			if val.CanFloat() {
				floatVal = val.Float()
			} else if val.CanInt() {
				floatVal = float64(val.Int())
			} else if val.CanUint() {
				floatVal = float64(val.Uint())
			} else {
				continue
			}

			metricName := "wombatt_" + strings.ToLower(re.ReplaceAllString(key, "_"))
			fmt.Fprintf(w, "%s{source=\"%s\"} %v\n", metricName, source, floatVal)
		}
	}
}

func filterFields(page map[string]any, fields string) map[string]any {
	if fields == "" {
		return page
	}
	newPage := make(map[string]any)
	for k := range strings.SplitSeq(fields, ",") {
		if v, ok := page[k]; ok {
			newPage[k] = v
		}
	}
	return newPage
}
