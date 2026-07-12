package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInternalPortNilCheck(t *testing.T) {
	p := &internalPort{
		ReadWriteCloser: nil,
	}

	buf := make([]byte, 10)
	n, err := p.Read(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err)
	assert.Equal(t, "port is closed", err.Error())

	n, err = p.Write(buf)
	assert.Equal(t, 0, n)
	assert.Error(t, err)
	assert.Equal(t, "port is closed", err.Error())

	err = p.Close()
	assert.NoError(t, err)
}

type dummyReadWriteCloser struct{}

func (dummyReadWriteCloser) Read(p []byte) (n int, err error) {
	return len(p), nil
}

func (dummyReadWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (dummyReadWriteCloser) Close() error {
	return nil
}

func TestInternalPortLockConcurrentIO(t *testing.T) {
	p := &internalPort{
		ReadWriteCloser: dummyReadWriteCloser{},
	}

	p.Lock()
	defer p.Unlock()

	// Spawn a goroutine to read and write. If there was a deadlock, this would hang.
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 10)
		_, _ = p.Read(buf)
		_, _ = p.Write(buf)
		close(done)
	}()

	select {
	case <-done:
		// Success!
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock detected! Read/Write blocked while Lock is held")
	}
}
