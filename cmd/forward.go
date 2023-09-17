package cmd

import (
	"encoding/hex"
	"log"
	"path/filepath"
	"time"

	"wombatt/internal/common"

	"go.bug.st/serial"
)

type ForwardCmd struct {
	Controller  string `name:"controller-port" required:"" help:"Serial port of the RS485 controller"`
	Subordinate string `name:"subordinate-port" required:"" help:"A subordinate device in the RS485 bus"`
	BaudRate    uint   `short:"B" default:"9600" help:"Baud rate"`
}

func (fc *ForwardCmd) Run(globals *Globals) error {
	f := NewForward(fc)
	if err := f.Init(); err != nil {
		log.Fatalf("error initializing: %v\n", err)
	}
	f.RunForever()
	return nil
}

type Forward struct {
	ForwardCmd

	controller  common.Port
	subordinate common.Port
}

func NewForward(fo *ForwardCmd) *Forward {
	return &Forward{ForwardCmd: *fo}
}

func (f *Forward) Init() error {
	opts := &common.PortOptions{
		Mode: &serial.Mode{BaudRate: int(f.BaudRate)},
	}
	opts.Name = f.Controller
	port, err := common.OpenPort(opts)
	if err != nil {
		return err
	}
	f.controller = port

	opts.Name = f.Subordinate
	port, err = common.OpenPort(opts)
	if err != nil {
		f.controller.Close()
		return err
	}
	f.subordinate = port
	return nil
}

func (f *Forward) RunForever() {
	read := func(p common.Port) ([]byte, error) {
		b := make([]byte, 128)
		n, err := p.Read(b)
		return b[0:n], err
	}

	reopenOnError := func(err error, p common.Port, d, op string) {
		log.Printf("error %s %v: %v\n", op, d, err)
		if err := p.ReopenWithBackoff(); err != nil {
			log.Printf("error reopening: %v\n", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	readWrite := func(from, to common.Port, fname, tname string) {
		data, err := read(from)
		if err != nil {
			reopenOnError(err, from, fname, "reading")
			return
		}
		log.Printf("%v: %d %s\n", fname, len(data), hex.EncodeToString(data))
		_, err = to.Write(data)
		if err != nil {
			reopenOnError(err, to, tname, "writing")
			return
		}
	}

	go func() {
		for {
			readWrite(f.controller, f.subordinate, filepath.Base(f.Controller), f.Subordinate)
		}
	}()

	for {
		readWrite(f.subordinate, f.controller, filepath.Base(f.Subordinate), f.Controller)
	}
}
