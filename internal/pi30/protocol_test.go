package pi30

import (
	"bytes"
	"context"
	"fmt"
	"os"
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
		commands   []string
		response   string
		nfields    int
		writeError bool
		errstr     []string
	}{
		{
			commands: []string{"Q1"},
			errstr:   []string{"EOF"},
		},
		{
			commands: []string{"Q1"},
			response: "\r",
			errstr:   []string{"wrong start"},
		},
		{
			commands: []string{"Q1"},
			response: "(0\r",
			errstr:   []string{"short response"},
		},
		{
			commands: []string{"Q1"},
			response: "(00001 00000 00 00 00 040 047 048 042 01 00 000 0031 0000 0000 60.02 10 0 060 030 120 030 58.40 000 120 0 0000\x1d\xeb\r",
		},
		{
			commands: []string{"QPIGS"},
			response: "(123.5 60.0 120.1 60.0 2616 2442 040 364 52.50 000 093 0049 00.0 000.0 00.00 00051 00010000 00 00 00000 010\x97\x3b\r",
			nfields:  21,
		},
		{
			commands: []string{"QPIRI"},
			response: "(120.0 54.1 120.0 60.0 54.1 6500 6500 48.0 46.0 45.0 56.0 56.0 2 010 120 1 2 2 9 01 0 7 54.0 0 1 480 0 120h=\r",
			nfields:  28,
		},
		{
			commands: []string{"QPGS1"},
			response: "(1 96342210107441 B 00 123.2 60.02 120.1 60.00 2160 2079 033 52.5 000 073 000.0 000 04838 04595 035 00000010 5 2 120 120 002 00 047 000.0 00\x6d\xd9\r",
			nfields:  29,
		},
		{
			commands: []string{"QT"},
			response: "(2023081415283\xae\xf1\r",
			nfields:  1,
		},
		{
			commands: []string{"QT"},
			response: "(2023081415283\x0e\xf1\r",
			errstr:   []string{"crc error"},
		},
		{
			commands: []string{"QPIGS2"},
			response: "(00.0 000.0 00000 \x45\x93\r",
			nfields:  3,
		},
		{
			commands:   []string{"QPIGS2"},
			response:   "(00.0 000.0 00000 E\x00\r",
			writeError: true,
			errstr:     []string{"write error"},
		},
		{
			commands: []string{"QTA"}, // made up command
			response: "(NAKss\r",
			errstr:   []string{"NAK received"},
		},
	}

	for ii, tt := range tests {
		rw := newReadWriter([]byte(tt.response), tt.writeError)
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
			common.WriteTo(os.Stdout, results[0])
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

func TestValid(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		response string
		valid    bool
	}{
		{
			response: "(000.0 00.0 000.0 00.0 0000 0000 000 369 52.60 000 068 0033 00.0 000.0 00.00 00000 01000000 00 00 00000 010\x88\x95\r",
		},
		{
			response: "(125.1 60.0 119.8 52.1 0000 0000 000 369 52.60 000 068 0033 00.0 000.0 00.00 00000 01000000 00 00 00000 010\x31\x76\r",
			valid:    true,
		},
		{
			response: "(000.0 00.0 119.8 52.1 0000 0000 000 369 52.60 000 068 0033 00.0 000.0 00.00 00000 01000000 00 00 00000 010\x33\xc1\r",
			valid:    true,
		},
		{
			response: "(125.1 60.0 000.0 00.0 0000 0000 000 369 52.60 000 068 0033 00.0 000.0 00.00 00000 01000000 00 00 00000 010\x8a\x22\r",
			valid:    true,
		},
	}

	for ii, tt := range tests {
		b := []byte(tt.response)
		rw := newReadWriter(b, false)
		result, err := RunCommand(ctx, rw, "QPIGS")
		if err != nil {
			t.Errorf("error (#%d): got error %v", ii, err)
			continue
		}
		v, ok := result.(ResponseChecker)
		if !ok {
			continue
		}
		if tt.valid != v.Valid() {
			t.Errorf("error (#%d): got %v; want %v", ii, v.Valid(), tt.valid)
		}
	}
}
