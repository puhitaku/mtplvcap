package mtp

import (
	"fmt"
	"io"
)

type Device interface {
	Configure() error
	RunTransactionWithNoParams(code uint16) error
	RunTransaction(req *Container, rep *Container, dest io.Writer, src io.Reader, writeSize int64) error
	GetDevicePropDesc(propCode uint16, info *DevicePropDesc) error
	GetDevicePropValue(propCode uint32, dest interface{}) error
	SetDevicePropValue(propCode uint32, src interface{}) error
	ID() (ID, error)
	// Connected reports whether the underlying USB device is still open and
	// usable. It returns false once the device has been unplugged (the
	// transaction layer closes the handle on a fatal USB error).
	Connected() bool
	// Close releases the interface and closes the device.
	Close() error
	// Done releases the underlying USB device reference. It must be called once
	// the device is no longer needed (after Close) to avoid leaking a reference
	// across reconnects.
	Done()
}

// Opener (re)establishes a connection to a matching MTP device. It keeps the
// long-lived USB context and the search criteria so it can be called
// repeatedly to reconnect after the device has been unplugged and plugged
// back in.
type Opener interface {
	// Open enumerates the bus, opens, claims and configures a matching MTP
	// device, returning a device that is ready for transactions.
	Open() (Device, error)
}

type sessionData struct {
	tid uint32
	sid uint32
}

// RCError are return codes from the Container.Code field.
type RCError uint16

func (e RCError) Error() string {
	n, ok := RC_names[int(e)]
	if ok {
		return n
	}
	return fmt.Sprintf("RetCode %x", uint16(e))
}

// SyncError is an error type that indicates lost transaction
// synchronization in the protocol.
type SyncError string

func (s SyncError) Error() string {
	return string(s)
}

type Catastrophic string

func (f Catastrophic) Error() string {
	return string(f)
}

// The linux usb stack can send 16kb per call, according to libusb.
const rwBufSize = 0x4000

type DebugFlags struct {
	MTP  bool
	USB  bool
	Data bool
}
