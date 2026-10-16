package common

import (
	"testing"

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
