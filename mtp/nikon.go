package mtp

import "strings"

// Nikon MTP extensions

const (
	OC_NIKON_AfDrive = 0x90C1
	OC_NIKON_DeviceReady = 0x90C8
	DPC_NIKON_RecordingMedia = 0xD10B
)

type Rotation int

const (
	Rotation0       Rotation = 0
	Rotation90      Rotation = 90
	RotationMinus90 Rotation = -90
	Rotation180     Rotation = 180
)

type AF int

const (
	AFNotActive AF = 0
	AFFail      AF = 1
	AFSuccess   AF = 2
)

type RecordingMedia int8

const (
	RecordingMediaCard RecordingMedia = 0
	RecordingMediaSDRAM = 1
)

type Model struct {
	Name string
	HeaderSize int
}

type ModelMap map[string]Model

func (mm ModelMap) Match(product string) (Model, bool) {
	for k, v := range models {
		if strings.Contains(product, k) {
			return v, true
		}
	}
	return Model{}, false
}

func (mm ModelMap) Generic() Model {
	return mm["_generic"]
}

var models = ModelMap{
	"_generic": {
		Name: "Generic",
		HeaderSize: 384,
	},
	"D5000": {
		Name: "D5000",
		HeaderSize: 128,
	},
	"D5300": {
		Name: "D5300",
		HeaderSize: 384,
	},
}
