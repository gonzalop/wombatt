package cmd

import (
	"fmt"
	"log/slog"

	"wombatt/internal/common"
	"wombatt/internal/modbus"
	"wombatt/internal/solark"
)

type SolarkQueryCmd struct {
	PortName string `short:"p" help:"Port name (e.g., /dev/ttyUSB0, COM1, tcp://localhost:502)" required:"true"`
	BaudRate int    `short:"B" help:"Baud rate for serial port" default:"9600"`
	DataBits int    `help:"Number of data bits for serial port" default:"8"`
	StopBits int    `help:"Number of stop bits for serial port" default:"1"`
	Parity   string `help:"Parity for serial port (N, E, O)" default:"N"`
	Protocol string `short:"R" help:"Modbus protocol (auto, ModbusRTU, ModbusTCP)" default:"auto"`
	ModbusID int    `short:"i" help:"Modbus slave ID" default:"1"`
}

func (cmd *SolarkQueryCmd) Run() error {
	if cmd.ModbusID < 0 || cmd.ModbusID > 255 {
		return fmt.Errorf("invalid modbus ID: %d", cmd.ModbusID)
	}
	port, err := common.NewPort(cmd.PortName, cmd.BaudRate, cmd.DataBits, cmd.StopBits, cmd.Parity)
	if err != nil {
		slog.Error("failed to create port", "error", err)
		return err
	}
	defer port.Close()

	reader, err := modbus.Reader(port, cmd.Protocol, "")
	if err != nil {
		slog.Error("failed to create modbus reader", "error", err)
		return err
	}

	rtd, err := solark.ReadRealtimeData(reader, uint8(cmd.ModbusID))
	if err != nil {
		slog.Error("failed to read realtime data", "error", err)
		return err
	}

	fmt.Println("--- Solark Real-time Data ---")
	common.TraverseStruct(rtd, func(tags map[string]string, value any) {
		fmt.Printf("%s: %v %s\n", tags["name"], value, tags["unit"])
	})

	ia, err := solark.ReadIntrinsicAttributes(reader, uint8(cmd.ModbusID))
	if err != nil {
		slog.Error("failed to read intrinsic attributes", "error", err)
		return err
	}

	fmt.Println("\n--- Solark Intrinsic Attributes ---")
	fmt.Printf("Serial Number: %s\n", ia.SerialNumber())

	return nil
}
