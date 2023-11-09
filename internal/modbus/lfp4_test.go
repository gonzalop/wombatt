package modbus

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"wombatt/internal/common"
)

type testPort struct {
	reader io.Reader
	writer io.Writer
	cb     CB
}

type CB func(b []byte)

// NewTestPort implements the common.Port interface
func NewTestPort(cb CB, r io.Reader, w io.Writer) *testPort {
	return &testPort{cb: cb, reader: r, writer: w}
}

func (tp *testPort) Read(b []byte) (n int, err error) {
	if tp.cb != nil {
		tp.cb(b)
	}
	return tp.reader.Read(b)
}

func (tp *testPort) Write(b []byte) (n int, err error) {
	if tp.cb != nil {
		tp.cb(b)
	}
	return tp.writer.Write(b)
}

func (*testPort) ReopenWithBackoff() error {
	return nil
}

func (*testPort) Type() common.DeviceType {
	return common.TestByteDevice

}

func (*testPort) Close() error {
	return nil
}

// test data from examples in https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
// TestLFP4Request test the raw requests content.
func TestLFP4Request(t *testing.T) {
	tests := []struct {
		id   uint8
		cid2 uint8
		req  string
	}{
		{
			id:   1,
			cid2: 0x42, // Get analog value, fixed point
			req:  "7e323030313441343230303030464441320d",
		},
		{
			id:   1,
			cid2: 0x44, // Get alarm information
			req:  "7e323030313441343430303030464441300d",
		},
	}
	for tid, tt := range tests {
		req, err := hex.DecodeString(tt.req)
		if err != nil {
			t.Fatalf("malformed request string in test %d: %s", tid, tt.req)
		}
		frame := buildReadRequestLFP4Frame(tt.id, tt.cid2)
		if !bytes.Equal(frame.RawData(), req) {
			t.Errorf("test %d got '%s'; want '%s'", tid, hex.EncodeToString(frame.RawData()), tt.req)
		}
	}
}

// test data from examples in https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
// TestLFP4Response tests the raw response contents before being processed by ReadRegisters.
func TestLFP4Response(t *testing.T) {
	tests := []struct {
		id   uint8
		cid2 uint8
		resp string
	}{
		{
			id:   1,
			cid2: 0x42, // Get analog value, fixed point
			resp: "7e32303031344130304130434130313031313030433534304338313043383130433832304338313043383130433831304338313043383230433832304338323043383230433832304338323043383230433745303430424344304243443042434430424344304244373042443730303030313346443030303032373130303030303046303030303030363430433832304335343030324530424344304243443030303030303135303030303030334330303030303030413030303030303144303030303030303030303030303030303030303230303144443330300d",
		},
		{
			id:   1,
			cid2: 0x44, // Get alarm information
			resp: "7e323030313441303037303534303130313130303030303030303030303030303030303030303030303030303030303030303030343030303030303030303030303030303030393030303030303030303030313033303030303030303030303030454443340d",
		},
	}
	for tid, tt := range tests {
		resp, err := hex.DecodeString(tt.resp)
		if err != nil {
			t.Fatalf("malformed response string in test %d: %s", tid, tt.resp)
		}
		port := NewTestPort(nil, bytes.NewReader(resp), io.Discard)
		reader, _ := ReaderFromProtocol(port, "lifepower4")
		defer reader.Close()
		lfp4, ok := reader.(*LFP4)
		if !ok {
			t.Fatalf("wrong reader type: got %T want *LFP4", lfp4)
		}
		frame, err := lfp4.ReadResponse(1)
		if err != nil {
			t.Errorf("test got error %v", err)
		} else if !bytes.Equal(frame.RawData(), resp) {
			t.Errorf("test %d got \n'%s'; want \n'%s'", tid, hex.EncodeToString(frame.RawData()), tt.resp)
		}
	}
}

// test data from examples in https://eg4electronics.com/backend/wp-content/uploads/2023/04/EG4_LifePower4_Communication_Protocol.pdf
// TestLFP4Response tests the raw response contents before being processed by ReadRegisters.
func TestLFP4ReadRegisters(t *testing.T) {
	tests := []struct {
		id       uint8
		cid2     uint8
		rawResp  string
		dataResp string
	}{
		{
			id:       1,
			cid2:     0x42, // Get analog value, fixed point
			rawResp:  "7e32303031344130304130434130313031313030433534304338313043383130433832304338313043383130433831304338313043383230433832304338323043383230433832304338323043383230433745303430424344304243443042434430424344304244373042443730303030313346443030303032373130303030303046303030303030363430433832304335343030324530424344304243443030303030303135303030303030334330303030303030413030303030303144303030303030303030303030303030303030303230303144443330300d",
			dataResp: "303130313130304335343043383130433831304338323043383130433831304338313043383130433832304338323043383230433832304338323043383230433832304337453034304243443042434430424344304243443042443730424437303030303133464430303030323731303030303030463030303030303634304338323043353430303245304243443042434430303030303031353030303030303343303030303030304130303030303031443030303030303030303030303030303030303032303031444433",
		},
		{
			id:       1,
			cid2:     0x44, // Get alarm information
			rawResp:  "7e323030313441303037303534303130313130303030303030303030303030303030303030303030303030303030303030303030343030303030303030303030303030303030393030303030303030303030313033303030303030303030303030454443340d",
			dataResp: "3031303131303030303030303030303030303030303030303030303030303030303030303030303430303030303030303030303030303030303930303030303030303030303130333030303030303030303030304544",
		},
	}
	for tid, tt := range tests {
		rawResp, err := hex.DecodeString(tt.rawResp)
		if err != nil {
			t.Fatalf("malformed raw response string in test %d: %s", tid, tt.rawResp)
		}
		port := NewTestPort(nil, bytes.NewReader(rawResp), io.Discard)
		lfp4, _ := ReaderFromProtocol(port, "lifepower4")
		defer lfp4.Close()
		// dataResp needs double decoding: one from the test data to rawData and one from that to actual binary data
		// which is what ReadRegisters returns.
		dataResp, err := hex.DecodeString(tt.dataResp)
		if err != nil {
			t.Fatalf("malformed data response string in test %d: %s", tid, tt.dataResp)
		}
		dataResp, err = hex.DecodeString(string(dataResp))
		if err != nil {
			t.Fatalf("malformed decoded data response string in test %d: %s", tid, tt.dataResp)
		}
		frame, err := lfp4.ReadRegisters(tt.id, 0, tt.cid2)
		if err != nil {
			t.Errorf("test got error %v", err)
		} else if !bytes.Equal(frame.RawData(), dataResp) {
			t.Errorf("test %d got \n'%s'; want \n'%s'", tid, hex.EncodeToString(frame.RawData()), hex.EncodeToString(dataResp))
		}
	}
}