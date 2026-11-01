package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"wombatt/internal/common"
)

type fakeRead struct {
	data []byte
	err  error
}

// fakePort plays back scripted reads, then reports idle timeouts.
type fakePort struct {
	mu        sync.Mutex
	reads     []fakeRead
	writeErrs []error
	written   bytes.Buffer
	reopens   int
	reopenErr error
	// blockReopen makes ReopenWithBackoff wait for ctx cancellation.
	blockReopen   bool
	reopenStarted chan struct{}
}

func (p *fakePort) Read(b []byte) (int, error) {
	p.mu.Lock()
	if len(p.reads) == 0 {
		p.mu.Unlock()
		time.Sleep(time.Millisecond)
		return 0, common.ErrReadTimeout
	}
	r := p.reads[0]
	p.reads = p.reads[1:]
	p.mu.Unlock()
	return copy(b, r.data), r.err
}

func (p *fakePort) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.writeErrs) > 0 {
		err := p.writeErrs[0]
		p.writeErrs = p.writeErrs[1:]
		if err != nil {
			return 0, err
		}
	}
	return p.written.Write(b)
}

func (p *fakePort) ReopenWithBackoff(ctx context.Context) error {
	p.mu.Lock()
	p.reopens++
	block, started := p.blockReopen, p.reopenStarted
	err := p.reopenErr
	p.mu.Unlock()
	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	return err
}

func (p *fakePort) state() (string, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.written.String(), p.reopens
}

func (*fakePort) ResetInputBuffer() error            { return nil }
func (*fakePort) SetReadTimeout(time.Duration) error { return nil }
func (*fakePort) Type() common.DeviceType            { return common.TestByteDevice }
func (*fakePort) Close() error                       { return nil }

// startForward runs ForwardCmd.run in the background and returns a stop
// function that cancels it and fails the test if it does not return promptly.
func startForward(t *testing.T, controller, subordinate common.Port) (context.CancelFunc, func()) {
	t.Helper()
	old := reopenPause
	reopenPause = time.Millisecond
	t.Cleanup(func() { reopenPause = old })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	cmd := &ForwardCmd{Controller: "/dev/ctrl", Subordinate: "/dev/sub"}
	go func() {
		cmd.run(ctx, controller, subordinate)
		close(done)
	}()
	stop := func() {
		t.Helper()
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("run did not return after cancel")
		}
	}
	return cancel, stop
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestForwardRun(t *testing.T) {
	tests := []struct {
		name            string
		ctrlReads       []fakeRead
		subWriteErrs    []error
		reopenErr       error
		wantSub         string
		wantCtrlReopens int
		wantSubReopens  int
	}{
		{
			name:      "forwards data",
			ctrlReads: []fakeRead{{data: []byte("abc")}},
			wantSub:   "abc",
		},
		{
			name:      "idle timeouts do not reopen",
			ctrlReads: []fakeRead{{err: common.ErrReadTimeout}, {err: common.ErrReadTimeout}, {data: []byte("x")}},
			wantSub:   "x",
		},
		{
			name:            "read error reopens and keeps forwarding",
			ctrlReads:       []fakeRead{{err: io.ErrUnexpectedEOF}, {data: []byte("y")}},
			wantSub:         "y",
			wantCtrlReopens: 1,
		},
		{
			name:            "reopen failure keeps looping",
			ctrlReads:       []fakeRead{{err: io.ErrUnexpectedEOF}, {data: []byte("z")}},
			reopenErr:       errors.New("no device"),
			wantSub:         "z",
			wantCtrlReopens: 1,
		},
		{
			name:            "data returned with an error is still forwarded",
			ctrlReads:       []fakeRead{{data: []byte("p"), err: io.ErrUnexpectedEOF}},
			wantSub:         "p",
			wantCtrlReopens: 1,
		},
		{
			name:           "write error reopens destination",
			ctrlReads:      []fakeRead{{data: []byte("lost")}, {data: []byte("w")}},
			subWriteErrs:   []error{errors.New("write failed")},
			wantSub:        "w",
			wantSubReopens: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := &fakePort{reads: tt.ctrlReads, reopenErr: tt.reopenErr}
			sub := &fakePort{writeErrs: tt.subWriteErrs}
			_, stop := startForward(t, ctrl, sub)
			waitFor(t, "forwarded data", func() bool {
				got, _ := sub.state()
				return got == tt.wantSub
			})
			waitFor(t, "controller reopens", func() bool {
				_, n := ctrl.state()
				return n == tt.wantCtrlReopens
			})
			stop()
			if _, n := ctrl.state(); n != tt.wantCtrlReopens {
				t.Errorf("controller reopens = %d, want %d", n, tt.wantCtrlReopens)
			}
			if _, n := sub.state(); n != tt.wantSubReopens {
				t.Errorf("subordinate reopens = %d, want %d", n, tt.wantSubReopens)
			}
		})
	}
}

func TestForwardRunCancelDuringReopen(t *testing.T) {
	ctrl := &fakePort{
		reads:         []fakeRead{{err: io.ErrUnexpectedEOF}},
		blockReopen:   true,
		reopenStarted: make(chan struct{}, 1),
	}
	sub := &fakePort{}
	_, stop := startForward(t, ctrl, sub)
	select {
	case <-ctrl.reopenStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("reopen never started")
	}
	stop() // fails if run blocks in reopen or in the post-reopen pause
}
