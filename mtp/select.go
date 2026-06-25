package mtp

import (
	"fmt"
	"strings"

	"github.com/google/gousb"
	"github.com/hanwen/usb"
)

func SelectDeviceGoUSB(ctx *gousb.Context, vid, pid uint16) (*DeviceGoUSB, error) {
	var mtpDev []*DeviceGoUSB

	if vid != 0 && pid != 0 {
		log.USB.Infof("searching %04d:%04d", vid, pid)
	}

	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		v, p := uint16(desc.Vendor), uint16(desc.Product)
		if vid != 0 && pid != 0 && (v != vid || p != pid) {
			return false
		}

		for _, conf := range desc.Configs {
			for _, iface := range conf.Interfaces {
				hasImageClass := false
				for _, alt := range iface.AltSettings {
					hasImageClass = hasImageClass || alt.Class == gousb.ClassPTP
				}
				if !hasImageClass {
					continue
				}

				for _, alt := range iface.AltSettings {
					if len(alt.Endpoints) != 3 {
						continue
					}

					var ev, fe, se gousb.EndpointDesc
					for _, ep := range alt.Endpoints {
						switch {
						case ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeInterrupt:
							ev = ep
						case ep.Direction == gousb.EndpointDirectionIn && ep.TransferType == gousb.TransferTypeBulk:
							fe = ep
						case ep.Direction == gousb.EndpointDirectionOut && ep.TransferType == gousb.TransferTypeBulk:
							se = ep
						}
					}

					if se.Address > 0 && fe.Address > 0 && ev.Address > 0 {
						d := &DeviceGoUSB{
							devDesc:     desc,
							ifaceDesc:   iface,
							sendEPDesc:  se,
							fetchEPDesc: fe,
							eventEPDesc: ev,
							configDesc:  conf,

							iConfiguration: conf.Number,
							iInterface:     iface.Number,
							iAltSetting:    alt.Number,
						}
						mtpDev = append(mtpDev, d)

						log.USB.Infof("found: %04x:%04x", v, p)
						return true
					}
				}
			}
		}
		return false
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate USB devices: %s", err)
	}

	if len(mtpDev) == 0 {
		return nil, fmt.Errorf("found no MTP devices")
	} else if len(mtpDev) > 1 {
		var s []string
		for i, d := range mtpDev {
			s = append(s, fmt.Sprintf("%d. %04x:%04x", i+1, d.devDesc.Vendor, d.devDesc.Product))
		}
		return nil, fmt.Errorf("found multiple MTP devices: %s", strings.Join(s, ", "))
	}

	found := mtpDev[0]
	found.dev = devs[0]
	return found, nil
}

func candidateFromDeviceDescriptor(d *usb.Device) *DeviceDirect {
	dd, err := d.GetDeviceDescriptor()
	if err != nil {
		return nil
	}
	for i := byte(0); i < dd.NumConfigurations; i++ {
		cdecs, err := d.GetConfigDescriptor(i)
		if err != nil {
			return nil
		}
		for _, iface := range cdecs.Interfaces {
			for _, a := range iface.AltSetting {
				if len(a.EndPoints) != 3 {
					continue
				}
				m := DeviceDirect{}
				for _, s := range a.EndPoints {
					switch {
					case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_INTERRUPT:
						m.eventEP = s.EndpointAddress
					case s.Direction() == usb.ENDPOINT_IN && s.TransferType() == usb.TRANSFER_TYPE_BULK:
						m.fetchEP = s.EndpointAddress
					case s.Direction() == usb.ENDPOINT_OUT && s.TransferType() == usb.TRANSFER_TYPE_BULK:
						m.sendEP = s.EndpointAddress
					}
				}
				if m.sendEP > 0 && m.fetchEP > 0 && m.eventEP > 0 {
					m.devDescr = *dd
					m.ifaceDescr = a
					m.dev = d.Ref()
					m.configValue = cdecs.ConfigurationValue
					return &m
				}
			}
		}
	}

	return nil
}

// DirectOpener enumerates and opens MTP devices via libusb (hanwen/usb). It
// keeps a single long-lived libusb context so the bus can be re-enumerated to
// reconnect after the device has been unplugged and plugged back in.
type DirectOpener struct {
	ctx *usb.Context
	vid uint16
	pid uint16

	// wantSerial holds the serial number of the device that was opened last
	// time. On a subsequent Open it is preferred over other matching devices so
	// we reconnect to the same physical camera. It is best-effort: if no device
	// with this serial is present, any matching device is adopted instead.
	wantSerial string
}

// NewDirectOpener creates a DirectOpener for the given VID/PID (0 matches any).
// The underlying libusb context lives for the whole lifetime of the opener and
// is reused across reconnects.
func NewDirectOpener(vid, pid uint16) *DirectOpener {
	return &DirectOpener{
		ctx: usb.NewContext(),
		vid: vid,
		pid: pid,
	}
}

// Open enumerates the bus and returns an opened, claimed and configured MTP
// device. It implements the Opener interface and may be called repeatedly to
// reconnect.
func (o *DirectOpener) Open() (Device, error) {
	return o.open(true)
}

// open enumerates matching MTP devices and adopts one. When configure is true
// the MTP session is opened (via DeviceDirect.Configure) before returning.
func (o *DirectOpener) open(configure bool) (*DeviceDirect, error) {
	if o.ctx == nil {
		o.ctx = usb.NewContext()
	}

	list, err := o.ctx.GetDeviceList()
	if err != nil {
		return nil, err
	}
	// DeviceList.Done() frees the list by indexing &list[0], which panics on an
	// empty slice. GetDeviceList returns an empty (non-nil) list when the bus
	// has no devices, which happens e.g. while the only camera is unplugged
	// during a reconnect, so guard against it.
	if len(list) > 0 {
		defer list.Done()
	}

	candidates := o.buildCandidates(list)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no MTP devices found")
	}

	// Every candidate holds an extra reference (see candidateFromDeviceDescriptor).
	// Make sure each one we do not adopt is unreferenced before we return.
	var kept *DeviceDirect
	defer func() {
		for _, c := range candidates {
			if c != kept {
				c.Done()
			}
		}
	}()

	// Pass 1: prefer the device we were connected to before (by serial).
	if o.wantSerial != "" {
		if dev := o.adopt(candidates, o.wantSerial, configure); dev != nil {
			kept = dev
			return dev, nil
		}
		log.MTP.Infof("previously-connected camera (serial %s) not found; adopting any matching MTP device", o.wantSerial)
	}

	// Pass 2: adopt the first matching device that opens successfully.
	if dev := o.adopt(candidates, "", configure); dev != nil {
		kept = dev
		return dev, nil
	}

	return nil, fmt.Errorf("found MTP devices but failed to open any of them")
}

// buildCandidates returns the matching MTP devices in the list, ordered so a
// Nikon DSLR (VID 0x04b0) comes first. Every returned device holds an extra
// reference and must be released with Done by the caller.
func (o *DirectOpener) buildCandidates(list usb.DeviceList) []*DeviceDirect {
	var nikon, others []*DeviceDirect
	for _, d := range list {
		v, err := d.GetDeviceDescriptor()
		if err != nil {
			continue
		}
		if o.vid != 0 && v.IdVendor != o.vid {
			continue
		}
		if o.pid != 0 && v.IdProduct != o.pid {
			continue
		}
		cand := candidateFromDeviceDescriptor(d)
		if cand == nil {
			continue
		}
		log.USB.Infof("found: %04x:%04x", v.IdVendor, v.IdProduct)
		if v.IdVendor == 0x04b0 {
			nikon = append(nikon, cand)
		} else {
			others = append(others, cand)
		}
	}
	if len(nikon)+len(others) > 1 {
		log.MTP.Warningf("detected more than 1 device")
	}
	return append(nikon, others...)
}

// adopt opens, validates and (optionally) configures candidates in order,
// returning the first one accepted. When wantSerial is non-empty only a device
// whose serial matches is accepted. Candidates that are opened but not adopted
// are closed again; their reference is released by the caller of open.
func (o *DirectOpener) adopt(candidates []*DeviceDirect, wantSerial string, configure bool) *DeviceDirect {
	for _, dev := range candidates {
		vendor, product := dev.devDescr.IdVendor, dev.devDescr.IdProduct

		if err := dev.Open(); err != nil {
			log.MTP.Warningf("could not open %04x:%04x: %s", vendor, product, err)
			// Open may fail after the handle was opened (e.g. detach/claim
			// failure), so close it to release the handle and reattach the
			// kernel driver. Close is a no-op if the handle was never opened.
			dev.Close()
			continue
		}

		// Make sure the device is set to the configuration that exposes the MTP
		// interface.
		if config, err := dev.h.GetConfiguration(); err != nil {
			log.MTP.Warningf("could not get configuration of %04x:%04x: %s", vendor, product, err)
			dev.Close()
			continue
		} else if config != dev.configValue {
			if err := dev.h.SetConfiguration(dev.configValue); err != nil {
				log.MTP.Warningf("could not set configuration of %04x:%04x: %s", vendor, product, err)
				dev.Close()
				continue
			}
		}

		dev.Timeout = 3000

		serial := ""
		if id, err := dev.ID(); err == nil {
			serial = id.SerialNumber
		}
		if wantSerial != "" && serial != wantSerial {
			// Not the camera we want to reconnect to; release and keep looking.
			dev.Close()
			continue
		}

		if configure {
			if err := dev.Configure(); err != nil {
				log.MTP.Warningf("could not configure %04x:%04x: %s", vendor, product, err)
				dev.Close()
				continue
			}
		}

		// Remember the serial so the next reconnect prefers the same body, but
		// keep the previously known serial if this device could not report one
		// (otherwise a transient ID() failure would permanently disable pinning).
		if serial != "" {
			o.wantSerial = serial
		}
		log.MTP.Infof("opened %04x:%04x", vendor, product)
		return dev
	}
	return nil
}

// SelectDeviceDirect returns an opened MTP device that matches the given
// pattern. It is kept for backwards compatibility (e.g. tests); new code
// should use DirectOpener so the connection can be re-established later. The
// returned device is opened and claimed but not configured.
func SelectDeviceDirect(vid, pid uint16) (*DeviceDirect, error) {
	return NewDirectOpener(vid, pid).open(false)
}
