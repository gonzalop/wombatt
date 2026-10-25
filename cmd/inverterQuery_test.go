package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestProcessQueryResults(t *testing.T) {
	tests := []struct {
		name       string
		dev        string
		commands   []string
		results    []any
		errs       []error
		wantErr    bool
		errContain string
	}{
		{
			name:       "nil results with multiple errors",
			dev:        "/dev/ttyUSB0",
			commands:   []string{"Q1", "QPIGS"},
			results:    nil,
			errs:       []error{errors.New("timeout 1"), errors.New("timeout 2")},
			wantErr:    true,
			errContain: "timeout 1\ntimeout 2",
		},
		{
			name:       "nil results with single error",
			dev:        "/dev/ttyUSB0",
			commands:   []string{"Q1"},
			results:    nil,
			errs:       []error{errors.New("connection refused")},
			wantErr:    true,
			errContain: "connection refused",
		},
		{
			name:       "results with per-command error",
			dev:        "/dev/ttyUSB0",
			commands:   []string{"Q1", "QPIGS"},
			results:    []any{struct{ Val int }{Val: 1}, nil},
			errs:       []error{nil, errors.New("read failed")},
			wantErr:    true,
			errContain: "error running QPIGS on port /dev/ttyUSB0: read failed",
		},
		{
			name:     "successful query with output",
			dev:      "/dev/ttyUSB0",
			commands: []string{"Q1"},
			results:  []any{struct{ Val int }{Val: 42}},
			errs:     []error{nil},
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := processQueryResults(tc.dev, tc.commands, tc.results, tc.errs, &buf)
			if (err != nil) != tc.wantErr {
				t.Fatalf("processQueryResults() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr && tc.errContain != "" {
				if !strings.Contains(err.Error(), tc.errContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContain)
				}
			}
		})
	}
}
