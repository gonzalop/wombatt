package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// blockingWriter stalls Write like a client that stops reading.
type blockingWriter struct {
	header  http.Header
	writing chan struct{}
	release chan struct{}
}

func (w *blockingWriter) Header() http.Header { return w.header }
func (w *blockingWriter) WriteHeader(int)     {}
func (w *blockingWriter) Write(b []byte) (int, error) {
	close(w.writing)
	<-w.release
	return len(b), nil
}

func TestServeMetricsSlowClientDoesNotBlockPublish(t *testing.T) {
	srv := NewServer("127.0.0.1:0", "/test")
	srv.Publish("m", &struct {
		Volts float64 `name:"volts"`
	}{Volts: 1.5})

	w := &blockingWriter{header: http.Header{}, writing: make(chan struct{}), release: make(chan struct{})}
	served := make(chan struct{})
	go func() {
		srv.ServeMetrics(w, httptest.NewRequest("GET", "/metrics", nil))
		close(served)
	}()
	<-w.writing

	published := make(chan struct{})
	go func() {
		srv.Publish("m", &struct {
			Volts float64 `name:"volts"`
		}{Volts: 2.5})
		close(published)
	}()
	select {
	case <-published:
	case <-time.After(2 * time.Second):
		t.Error("Publish blocked while a metrics client was stalled")
	}
	close(w.release)
	<-served
}
