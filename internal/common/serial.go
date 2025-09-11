package common

import (
	"fmt"
	"io"
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

// DeviceType represents the type of communication device.
type DeviceType int

const (
	// TestByteDevice is a device type used for testing purposes.
	TestByteDevice DeviceType = iota
	// SerialDevice represents a serial port device.
	SerialDevice
	// HidRawDevice represents a HID raw device.
	HidRawDevice
	// TCPDevice represents a TCP network device.
	TCPDevice

	DefaultMaxBackoffInterval = 20 * time.Second
)

// DeviceTypeFromString maps string representations of device types to their DeviceType constants.
var DeviceTypeFromString = map[string]DeviceType{
	"test":   TestByteDevice,
	"serial": SerialDevice,
	"hidraw": HidRawDevice,
	"tcp":    TCPDevice,
}

// PortOptions contains the port name and the settings used when opening it.
// PortOptions contains the necessary parameters for opening a communication port.
type PortOptions struct {
	*serial.Mode // Mode contains serial port settings like BaudRate, DataBits, etc.

	Type    DeviceType // Type specifies the kind of device (e.g., SerialDevice, TCPDevice).
	Address string     // Address is the port name (e.g., "/dev/ttyUSB0") or network address (e.g., "192.168.1.1:8080").
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

// openSerial opens a serial port based on the provided options.
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

// openHidRaw opens a HID raw device (e.g., a file representing the device) based on the provided options.
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

// openTCP opens a TCP connection based on the provided options.
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
// Port defines the interface for a communication port (e.g., serial, TCP, HID).
type Port interface {
	io.ReadWriteCloser
	// ReopenWithBackoff attempts to close and then reopen the port with exponential backoff.
	// It returns an error if the port cannot be reopened after multiple retries.
	ReopenWithBackoff() error
	// Type returns the DeviceType of the port.
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

// OpenPortWithBackoff will keep trying to successfully open a new port for up to the specified duration.
func OpenPortWithBackoff(opts *PortOptions, d time.Duration) (Port, error) {
	f := func() (Port, error) {
		return OpenPort(opts)
	}
	n := func(err error, d time.Duration) {
		slog.Debug("backing off after error", "address", opts.Address, "error", err, "elapsed", d)
	}

	b := backoff.NewExponentialBackOff()
	if d == 0 {
		b.MaxElapsedTime = DefaultMaxBackoffInterval * 2 // Ensure at least one retry
	} else {
		b.MaxElapsedTime = d
	}
	b.MaxInterval = DefaultMaxBackoffInterval
	port, err := backoff.RetryNotifyWithData[Port](f, b, n)
	if err != nil {
		return port, err
	}
	return port, nil
}

func (p *internalPort) Type() DeviceType {
	return p.PortOptions.Type
}

// ReopenWithBackoff attempts to close the current port and then re-open it
// with an exponential backoff strategy until it succeeds.
func (p *internalPort) ReopenWithBackoff() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.Close()
	port, err := OpenPortWithBackoff(p.PortOptions, 0)
	p.ReadWriteCloser = port
	return err
}

// NewPort creates a new Port instance with the specified parameters.
// It handles serial port settings (baud rate, data bits, stop bits, parity)
// and determines the device type (serial, TCP, HID raw).
// It returns the initialized Port and an error if any parameter is invalid.
func NewPort(portName, deviceType string, baudRate, dataBits, stopBits int, parity string) (Port, error) {
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
