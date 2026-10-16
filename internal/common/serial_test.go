package common_test

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"wombatt/internal/common"

	"github.com/stretchr/testify/assert"
)

func TestNewTestPort(t *testing.T) {
	reader := bytes.NewBufferString("test read")
	writer := &bytes.Buffer{}
	dType := common.TestByteDevice

	port := common.NewTestPort(reader, writer, dType)

	assert.NotNil(t, port)
	assert.Equal(t, dType, port.Type())

	// Test Read
	buf := make([]byte, 9)
	n, err := port.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.Equal(t, "test read", string(buf))

	// Test Write
	n, err = port.Write([]byte("test write"))
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "test write", writer.String())
}

func TestNewPort_InvalidParity(t *testing.T) {
	_, err := common.NewPort("/dev/ttyUSB0", "serial", 9600, 8, 1, "X")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parity: X")
}

func TestNewPort_InvalidDeviceType(t *testing.T) {
	_, err := common.NewPort("/dev/ttyUSB0", "invalid_type", 9600, 8, 1, "N")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid device type: invalid_type")
}

func TestOpenPort_TestByteDevicePanics(t *testing.T) {
	opts := &common.PortOptions{
		Type: common.TestByteDevice,
	}
	assert.Panics(t, func() {
		_, err := common.OpenPort(opts)
		assert.Error(t, err)
	}, "OpenPort with TestByteDevice should panic")
}

func TestOpenPort_SerialDeviceFails(t *testing.T) {
	opts := &common.PortOptions{
		Type:    common.SerialDevice,
		Address: "/dev/nonexistent",
	}
	_, err := common.OpenPort(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening '/dev/nonexistent':")
}

func TestOpenPort_HidRawDeviceFails(t *testing.T) {
	opts := &common.PortOptions{
		Type:    common.HidRawDevice,
		Address: "/dev/nonexistent_hidraw",
	}
	_, err := common.OpenPort(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening '/dev/nonexistent_hidraw':")
}

func TestOpenPort_TCPDeviceFails(t *testing.T) {
	opts := &common.PortOptions{
		Type:    common.TCPDevice,
		Address: "localhost:12345",
	}
	_, err := common.OpenPort(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening 'localhost:12345':")
}

func TestNewPort_SerialDeviceFails(t *testing.T) {
	_, err := common.NewPort("/dev/nonexistent", "serial", 9600, 8, 1, "N")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening '/dev/nonexistent':")
}

func TestNewPort_HidRawDeviceFails(t *testing.T) {
	_, err := common.NewPort("/dev/nonexistent_hidraw", "hidraw", 9600, 8, 1, "N")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening '/dev/nonexistent_hidraw':")
}

func TestNewPort_TCPDeviceFails(t *testing.T) {
	_, err := common.NewPort("localhost:12345", "tcp", 9600, 8, 1, "N")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error opening 'localhost:12345':")
}

func TestOpenPortWithBackoffStopsOnCancel(t *testing.T) {
	// Grab a free port and close it so dialing it fails right away.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	port, err := common.OpenPortWithBackoff(ctx, &common.PortOptions{Type: common.TCPDevice, Address: addr}, 0)
	assert.Nil(t, port)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
}
