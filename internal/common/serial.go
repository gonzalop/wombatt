package common

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.bug.st/serial"
)

type DeviceType int

const (
	SerialDevice DeviceType = iota
	HidRawDevice
	TCPDevice
)

var DeviceTypeFromString = map[string]DeviceType{
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
	SerialDevice: openSerial,
	HidRawDevice: openHidRaw,
	TCPDevice:    openTCP,
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
	f, err := os.OpenFile(opts.Address, os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	o := *opts
	return &internalPort{ReadWriteCloser: f, PortOptions: &o}, nil
}

func openTCP(opts *PortOptions) (Port, error) {
	log.Printf("Opening %s...\n", opts.Address)
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
		log.Printf("error opening '%s': %v (%v elapsed)", opts.Address, err, d)
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
