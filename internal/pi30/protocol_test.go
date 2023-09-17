package pi30

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"wombatt/internal/common"
)

type readWriter struct {
	in         *bytes.Buffer
	writeError bool
}

func (rw *readWriter) Read(p []byte) (n int, err error) {
	return rw.in.Read(p)
}

func (rw *readWriter) Write(p []byte) (n int, err error) {
	if rw.writeError {
		return 0, fmt.Errorf("write error")
	}
	return len(p), nil
}

func newReadWriter(input []byte, writeError bool) *readWriter {
	return &readWriter{
		in:         bytes.NewBuffer(input),
		writeError: writeError,
	}
}

func TestRunCommands(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		commands    []string
		responseHex string
		nfields     int
		writeError  bool
		errstr      []string
	}{
		{
			commands: []string{"Q1"},
			errstr:   []string{"EOF"},
		},
		{
			commands:    []string{"Q1"},
			responseHex: "0d",
			errstr:      []string{"wrong start"},
		},
		{
			commands:    []string{"Q1"},
			responseHex: "28300d",
			errstr:      []string{"short response"},
		},
		{
			commands:    []string{"Q1"},
			responseHex: "28303030303120303030303020303020303020303020303430203034372030343820303432203031203030203030302030303331203030303020303030302036302e30322031302030203036302030333020313230203033302035382e34302030303020313230203020303030301deb0d",
		},
		{
			commands:    []string{"QPIGS"},
			responseHex: "283132332e352036302e30203132302e312036302e302032363136203234343220303430203336342035322e3530203030302030393320303034392030302e30203030302e302030302e303020303030353120303030313030303020303020303020303030303020303130973b0d",
			nfields:     21,
		},
		{
			commands:    []string{"QPIRI"},
			responseHex: "283132302e302035342e31203132302e302036302e302035342e31203635303020363530302034382e302034362e302034352e302035362e302035362e30203220303130203132302031203220322039203031203020372035342e302030203120343830203020313230683d0d",
			nfields:     28,
		},
		{
			commands:    []string{"QPGS1"},
			responseHex: "28312039363334323231303130373434312042203030203132332e322036302e3032203132302e312036302e303020323136302032303739203033332035322e352030303020303733203030302e3020303030203034383338203034353935203033352030303030303031302035203220313230203132302030303220303020303437203030302e302030306dd90d",
			nfields:     29,
		},
		{
			commands:    []string{"QT"},
			responseHex: "283230323330383134313532383334d3f30d",
			nfields:     1,
		},
		{
			commands:    []string{"QT"},
			responseHex: "283230323330383134313532383334d3f10d",
			errstr:      []string{"crc error"},
		},
		{
			commands:    []string{"QPIGS2"},
			responseHex: "2830302e30203030302e302030303030302045930d",
			nfields:     3,
		},
		{
			commands:    []string{"QPIGS2"},
			responseHex: "2830302e30203030302e302030303030302045930d",
			writeError:  true,
			errstr:      []string{"write error"},
		},
		{
			commands:    []string{"QTA"}, // made up command
			responseHex: "284e414b73730d",
			errstr:      []string{"NAK received"},
		},
	}

	for ii, tt := range tests {
		b, err := hex.DecodeString(tt.responseHex)
		if err != nil {
			t.Fatalf("error (%d) decoding string (%s): %v", ii, tt.responseHex, err)
		}
		rw := newReadWriter(b, tt.writeError)
		results, errors := RunCommands(ctx, rw, tt.commands)
		for i, err := range errors {
			if err != nil && (len(tt.errstr) == 0 || !strings.Contains(fmt.Sprintf("%v", err), tt.errstr[i])) {
				var e string
				if len(tt.errstr) != 0 {
					e = tt.errstr[0]
				}
				t.Errorf("error (#%d): '%v' '%s'", ii, err, e)
			}
			if err != nil && (len(tt.errstr) == 0 || tt.errstr[i] == "") {
				t.Errorf("error (#%d): got error '%v'; want no error", ii, err)
			}
			if err == nil && (len(tt.errstr) > 0 && tt.errstr[i] != "") {
				t.Errorf("error (#%d): got no error; want ~%s", ii, tt.errstr[i])
			}
		}
		if tt.nfields != 0 && results[0] != nil {
			WriteTo(io.Discard, results[0])
			counter := 0
			fu := func(map[string]string, any) {
				counter++
			}
			common.TraverseStruct(results[0], fu)
			if tt.nfields != counter {
				t.Errorf("wrong number of fields (#%d): got %d; want %d", ii, counter, tt.nfields)
			}
		}
	}
}
