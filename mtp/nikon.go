package mtp

// Nikon MTP extensions

const (
	OC_NIKON_AfDrive = 0x90C1
	OC_NIKON_DeviceReady = 0x90C8
	DPC_NIKON_RecordingMedia = 0xD10B
)

const (
	LVHeaderSize = 384
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