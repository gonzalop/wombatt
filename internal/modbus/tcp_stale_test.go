package modbus

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"wombatt/internal/common"
)

// TestTCPDropsLateReply checks that bytes left over from an earlier request
// that timed out are dropped before the next request is read.
func TestTCPDropsLateReply(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// A late reply with a transaction ID nobody is waiting for.
		_, _ = conn.Write([]byte{0xbe, 0xef, 0, 0, 0, 5, 1, 3, 2, 0, 9})
		req := make([]byte, 12)
		if _, err := io.ReadFull(conn, req); err != nil {
			return
		}
		var resp bytes.Buffer
		resp.Write(req[0:2])           // same transaction ID
		resp.Write([]byte{0, 0, 0, 5}) // protocol ID, length
		resp.Write([]byte{1, 3, 2})    // unit ID, function, byte count
		_ = binary.Write(&resp, binary.BigEndian, uint16(0x1234))
		_, _ = conn.Write(resp.Bytes())
	}()

	port, err := common.OpenPort(&common.PortOptions{Type: common.TCPDevice, Address: ln.Addr().String(), ReadTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer port.Close()
	time.Sleep(50 * time.Millisecond) // Let the late reply arrive.

	data, err := NewTCP(port).ReadHoldingRegisters(1, 0, 1)
	if err != nil {
		t.Fatalf("got error %v, want the fresh reply", err)
	}
	if !bytes.Equal(data, []byte{0x12, 0x34}) {
		t.Fatalf("got %x, want 1234", data)
	}
}
