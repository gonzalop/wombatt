package common

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.bug.st/serial"
)

type TestPort struct {
	reader io.Reader
	writer io.Writer
	dtype  DeviceType
}

// NewTestPort implements the Port interface
func NewTestPort(r io.Reader, w io.Writer, dtype DeviceType) *TestPort {
	return &TestPort{reader: r, writer: w, dtype: dtype}
}

func (tp *TestPort) Read(b []byte) (n int, err error) {
	return tp.reader.Read(b)
}

func (tp *TestPort) Write(b []byte) (n int, err error) {
	return tp.writer.Write(b)
}

func (*TestPort) ReopenWithBackoff() error {
	return nil
}

func (p *TestPort) Type() DeviceType {
	return p.dtype

}

func (*TestPort) Close() error {
	return nil
}

type DeviceType int

const (
	TestByteDevice DeviceType = iota
	SerialDevice
	HidRawDevice
	TCPDevice
)

var DeviceTypeFromString = map[string]DeviceType{
	"test":   TestByteDevice,
	"serial": SerialDevice,
	"hidraw": HidRawDevice,
	"tcp":    TCPDevice,
}

// PortOptions contains the port name and the settings used when opening it.
type PortOptions struct {
	*serial.Mode

	Type    DeviceType
	Address string
}

var deviceOpen = map[DeviceType]func(*PortOptions) (Port, error){
	TestByteDevice: func(*PortOptions) (Port, error) { panic("can't open a test device") },
	SerialDevice:   openSerial,
	HidRawDevice:   openHidRaw,
	TCPDevice:      openTCP,
}

type internalPort struct {
	io.ReadWriteCloser
	*PortOptions

	lock sync.Mutex
}

func openSerial(opts *PortOptions) (Port, error) {
	p, err := serial.Open(opts.Address, opts.Mode)
	if err != nil {
		return nil, err
	}
	_ = p.ResetInputBuffer()
	_ = p.ResetOutputBuffer()
	o := *opts
	return &internalPort{ReadWriteCloser: p, PortOptions: &o}, nil
}

func openHidRaw(opts *PortOptions) (Port, error) {
	// TODO: maybe try to emulate the baud rate from opts?
	slog.Debug("opening file", "file", opts.Address)
	f, err := os.OpenFile(opts.Address, os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	o := *opts
	return &internalPort{ReadWriteCloser: f, PortOptions: &o}, nil
}

func openTCP(opts *PortOptions) (Port, error) {
	slog.Debug("dialing TCP server", "address", opts.Address)
	conn, err := net.Dial("tcp", opts.Address)
	if err != nil {
		return nil, err
	}
	o := *opts
	return &internalPort{ReadWriteCloser: conn, PortOptions: &o}, nil
}

// Port adds one more functions opening a port with retries and exponential backoff.
type Port interface {
	io.ReadWriteCloser
	ReopenWithBackoff() error
	Type() DeviceType
}

// OpenPort opens a device.
func OpenPort(opts *PortOptions) (Port, error) {
	open := deviceOpen[opts.Type]
	if open == nil {
		return nil, fmt.Errorf("invalid device type: %v", opts.Type)
	}
	c, err := open(opts)
	if err != nil {
		return nil, fmt.Errorf("error opening '%s': %v", opts.Address, err)
	}
	return c, nil
}

// OpenPortOrFatal will terminate the program with an error message if it can't open the requested port.
func OpenPortOrFatal(opts *PortOptions) Port {
	res, err := OpenPort(opts)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return res
}

// OpenPortWithBackoff will keep trying to successfully open a new port for up to the specified duration.
func OpenPortWithBackoff(opts *PortOptions, d time.Duration) (Port, error) {
	f := func() (Port, error) {
		return OpenPort(opts)
	}
	n := func(err error, d time.Duration) {
		slog.Debug("backing off after error", "address", opts.Address, "error", err, "elapsed", d)
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = d
	b.MaxInterval = 20 * time.Second
	port, err := backoff.RetryNotifyWithData[Port](f, b, n)
	if err != nil {
		return port, err
	}
	return port, nil
}

func (p *internalPort) Type() DeviceType {
	return p.PortOptions.Type
}

// ReopenWithBackoff will close the port and forever try to open it until it succeeds.
func (p *internalPort) ReopenWithBackoff() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Close()
	port, err := OpenPortWithBackoff(p.PortOptions, 0)
	p.ReadWriteCloser = port
	return err
}

// NewPort creates a new Port based on the provided parameters.
func NewPort(portName string, baudRate, dataBits, stopBits int, parity string, deviceType string) (Port, error) {
	var serialParity serial.Parity
	switch strings.ToUpper(parity) {
	case "N":
		serialParity = serial.NoParity
	case "E":
		serialParity = serial.EvenParity
	case "O":
		serialParity = serial.OddParity
	default:
		return nil, fmt.Errorf("invalid parity: %s", parity)
	}

	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: serial.StopBits(stopBits),
		Parity:   serialParity,
	}

	portType, ok := DeviceTypeFromString[strings.ToLower(deviceType)]
	if !ok {
		return nil, fmt.Errorf("invalid device type: %s", deviceType)
	}
	address := portName

	opts := &PortOptions{
		Mode:    mode,
		Type:    portType,
		Address: address,
	}

	return OpenPort(opts)
}
