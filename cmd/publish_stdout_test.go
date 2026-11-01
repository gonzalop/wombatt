package cmd

import (
	"errors"
	"testing"

	"wombatt/internal/eg4_18kpv"
)

func TestPublishToStdoutSkipsFailedCommands(t *testing.T) {
	m := &inverterMonitor{Device: "test", Commands: []string{"RealtimeData", "RealtimeData"}}
	tests := []struct {
		name    string
		results []any
		errs    []error
	}{
		{"typed nil with error", []any{(*eg4_18kpv.RealtimeData)(nil), nil}, []error{errors.New("read failed"), errors.New("x")}},
		{"untyped nil", []any{nil, nil}, []error{nil, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publishToStdout(m, tt.results, tt.errs) // must not panic
		})
	}
}
