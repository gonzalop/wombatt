package modbus

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"wombatt/internal/common"
)

type typedNilResult struct{ A uint16 }

// A reader that returns a typed nil pointer on error must not leak it as a
// non-nil `any`: publishers check `result == nil` and would otherwise panic.
func TestRunCommandsDropsTypedNilOnError(t *testing.T) {
	port := common.NewTestPort(bytes.NewReader(nil), io.Discard, common.TestByteDevice)
	results, errs := RunCommands(context.Background(), port, RTUProtocol, 1, []string{"x"},
		func(RegisterReader, uint8, string) (any, error) {
			var p *typedNilResult
			return p, errors.New("boom")
		})
	if len(results) != 1 || len(errs) != 1 {
		t.Fatalf("got %d results, %d errors; want 1, 1", len(results), len(errs))
	}
	if results[0] != nil {
		t.Errorf("results[0] = %#v, want untyped nil", results[0])
	}
	if errs[0] == nil {
		t.Error("errs[0] = nil, want error")
	}
}
