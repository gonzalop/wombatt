package cmd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"wombatt/internal/pi30"
	"wombatt/internal/web"
)

// TestPublishClearsStaleDataOnFailure checks that a failure that returns no
// results (e.g. the port could not be opened) removes the old web page.
func TestPublishClearsStaleDataOnFailure(t *testing.T) {
	srv := web.NewServer("", "/inverter/")
	m := &inverterMonitor{Device: "/dev/x", Commands: []string{"QPIRI", "Q1"}, webServer: srv}
	status := func(path string) int {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}

	newCmdResponse(m, []any{&pi30.QPIRIResponse{}, &pi30.Q1Response{}}, []error{nil, nil}).Publish(context.Background(), "", 0)
	if got := status("/inverter/1/QPIRI"); got != http.StatusOK {
		t.Fatalf("after good poll: got status %d, want 200", got)
	}

	newCmdResponse(m, nil, []error{errors.New("no such device")}).Publish(context.Background(), "", 0)
	for _, path := range []string{"/inverter/1/QPIRI", "/inverter/1/Q1"} {
		if got := status(path); got != http.StatusNotFound {
			t.Errorf("after failed open: %s got status %d, want 404", path, got)
		}
	}
}

func TestNewCmdResponsePadsErrors(t *testing.T) {
	m := &inverterMonitor{Commands: []string{"A", "B"}}
	r := newCmdResponse(m, nil, nil)
	if len(r.Responses) != 2 || len(r.Errors) != 2 || r.Errors[0] == nil || r.Errors[1] == nil {
		t.Fatalf("got %d responses, errors %v; want 2 nil responses with errors", len(r.Responses), r.Errors)
	}
}
