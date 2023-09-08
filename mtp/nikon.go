package mtp

import "strings"

// Nikon MTP extensions

const (
	OC_NIKON_AfDrive         = 0x90C1
	OC_NIKON_DeviceReady     = 0x90C8
	DPC_NIKON_RecordingMedia = 0xD10B
	DPC_NIKON_Resolution     = 0xD1AC
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
type Resolution64 uint64
type Resolution8 uint8

type ResolutionType int

const (
	ResolutionType8 ResolutionType = iota
	ResolutionType64
)

const (
	RecordingMediaCard  RecordingMedia = 0
	RecordingMediaSDRAM                = 1
)

type Model struct {
	Name             string
	HeaderSize       int
	ResolutionType   ResolutionType
	QuirkSwitchMedia bool
	QuirkNoHeader    bool
}

type ModelMap map[string]Model

func (mm ModelMap) Match(product string) (Model, bool) {
	tokens := strings.Split(product, " ")
	for i := range tokens {
		tokens[i] = strings.ToLower(tokens[i])
	}

	for k, v := range models {
		for _, t := range tokens {
			if strings.ToLower(k) == t {
				return v, true
			}
		}
	}
	return Model{}, false
}

func (mm ModelMap) Generic() Model {
	return mm["_generic"]
}

var models = ModelMap{
	"_generic": {
		Name:       "Generic",
		HeaderSize: 384,
	},
	"D3": {
		Name:             "D3",
		HeaderSize:       128,
		QuirkSwitchMedia: true,
	},
	"D3s": {
		Name:             "D3s",
		HeaderSize:       128,
		QuirkSwitchMedia: true,
	},
	"D3X": {
		Name:             "D3X",
		HeaderSize:       64,
		QuirkSwitchMedia: true,
	},
	"D300": {
		Name:             "D300",
		HeaderSize:       64,
		QuirkSwitchMedia: true,
	},
	"D300s": {
		Name:             "D300s",
		HeaderSize:       64,
		QuirkSwitchMedia: true,
	},
	"D3200": {
		Name:       "D3200",
		HeaderSize: 384,
	},
	"D3300": {
		Name:       "D3300",
		HeaderSize: 384,
	},
	"D5000": {
		Name:             "D5000",
		HeaderSize:       128,
		QuirkSwitchMedia: true,
	},
	"D5300": {
		Name:       "D5300",
		HeaderSize: 384,
	},
	"D5500": {
		Name:       "D5500",
		HeaderSize: 384,
	},
	"D5600": {
		Name:       "D5600",
		HeaderSize: 384,
	},
	"D6": {
		Name:           "D6",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"D600": {
		Name:       "D600",
		HeaderSize: 384,
	},
	"D610": {
		Name:       "D610",
		HeaderSize: 384,
	},
	"D700": {
		Name:          "D700",
		HeaderSize:    64,
		QuirkNoHeader: true,
	},
	"D750": {
		Name:       "D750",
		HeaderSize: 384,
	},
	"D780": {
		Name:           "D780",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"D7000": {
		Name:       "D7000",
		HeaderSize: 384,
	},
	"D7200": {
		Name:       "D7200",
		HeaderSize: 384,
	},
	"D90": {
		Name:             "D90",
		HeaderSize:       128,
		QuirkSwitchMedia: true,
	},
	"6": {
		Name:           "Z6",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"6_2": {
		Name:           "Z6II",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"7": {
		Name:           "Z7",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"7_2": {
		Name:           "Z7II",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"9": {
		Name:           "Z9",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"50": {
		Name:           "Z50",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
	"fc": {
		Name:           "Zfc",
		HeaderSize:     384,
		ResolutionType: ResolutionType8,
	},
}
