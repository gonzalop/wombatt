//go:build linux

package common_test

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"wombatt/internal/common"
)

// mkfifo returns a FIFO path. Opened O_RDWR it behaves like a silent hidraw
// device: pollable, and Read blocks until data arrives.
func mkfifo(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hidraw")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skipf("mkfifo: %v", err)
	}
	return path
}

func TestHidRawReadTimesOut(t *testing.T) {
	tests := []struct {
		name        string
		openTimeout time.Duration
		setTimeout  time.Duration // applied via SetReadTimeout when > 0
	}{
		{"timeout from options", 100 * time.Millisecond, 0},
		{"timeout from SetReadTimeout", 0, 100 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := common.OpenPort(&common.PortOptions{
				Type:        common.HidRawDevice,
				Address:     mkfifo(t),
				ReadTimeout: tt.openTimeout,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer port.Close()
			if tt.setTimeout > 0 {
				assert.NoError(t, port.SetReadTimeout(tt.setTimeout))
			}

			done := make(chan error, 1)
			go func() {
				_, err := port.Read(make([]byte, 8))
				done <- err
			}()
			select {
			case err := <-done:
				assert.True(t, common.IsTimeout(err), "want timeout, got %v", err)
			case <-time.After(3 * time.Second):
				t.Fatal("hidraw Read on a silent device did not time out")
			}
		})
	}
}

func TestHidRawReadsData(t *testing.T) {
	port, err := common.OpenPort(&common.PortOptions{
		Type:        common.HidRawDevice,
		Address:     mkfifo(t),
		ReadTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer port.Close()
	_, err = port.Write([]byte("ok"))
	assert.NoError(t, err)
	buf := make([]byte, 8)
	n, err := port.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, "ok", string(buf[:n]))
}
