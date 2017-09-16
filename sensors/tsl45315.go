package sensors

import (
	"errors"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

const tsl45315I2CADDR = 0x29

type TSL45315Driver struct {
	name       string
	connector  i2c.Connector
	connection i2c.Connection
}

// Name returns the name for this Driver
func (t *TSL45315Driver) Name() string { return t.name }

// SetName sets the name for this Driver
func (t *TSL45315Driver) SetName(n string) { t.name = n }

// Connection returns the connection for this Driver
func (t *TSL45315Driver) Connection() gobot.Connection { return t.connector.(gobot.Connection) }

// Halt returns true if devices is halted successfully
func (t *TSL45315Driver) Halt() (err error) { return }

// Start initializes the TSL45315
func (t *TSL45315Driver) Start() (err error) {
	t.connection, err = t.connector.GetConnection(tsl45315I2CADDR, t.connector.GetDefaultBus())
	if err != nil {
		return
	}

	if _, err := t.connection.Write([]byte{0x80, 0x03}); err != nil {
		return err
	}

	if _, err := t.connection.Write([]byte{0x81, 0x02}); err != nil {
		return err
	}

	return
}

// Lux returns the current temperature, in celsius degrees.
func (t *TSL45315Driver) Lux() (lux float32, err error) {
	if _, err := t.connection.Write([]byte{0x84}); err != nil {
		return 0, err
	}
	time.Sleep(10 * time.Millisecond)
	data := make([]byte, 2)
	i, err := t.connection.Read(data)
	if err != nil {
		return
	}
	if i != 2 {
		return 0, errors.New("too few bytes")
	}

	lux = float32(4 * (uint32(data[1])<<8 | uint32(data[0])))

	return
}

func (t *TSL45315Driver) LuxTimes(times int) (lux float32, err error) {
	lux, err = t.Lux()
	if err != nil {
		return
	}

	var tempLux float32
	for i := 1; i < times; i++ {
		tempLux, err = t.Lux()
		if err != nil {
			return
		}
		lux = ((lux * float32(i)) + tempLux) / float32(i+1)
		time.Sleep(120 * time.Millisecond)
	}

	return
}

// NewTSL45315Driver creates a new driver
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
func NewTSL45315Driver(a i2c.Connector) *TSL45315Driver {
	s := &TSL45315Driver{
		name:      gobot.DefaultName("TSL45315"),
		connector: a,
	}

	return s
}
