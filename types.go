package scarlettctl

// ControlType represents the type of an ALSA control element
type ControlType int

const (
	ControlTypeNone ControlType = iota
	ControlTypeBoolean
	ControlTypeInteger
	ControlTypeEnumerated
	ControlTypeBytes
	ControlTypeIEC958
	ControlTypeInteger64
)

func (t ControlType) String() string {
	switch t {
	case ControlTypeBoolean:
		return "Boolean"
	case ControlTypeInteger:
		return "Integer"
	case ControlTypeEnumerated:
		return "Enumerated"
	case ControlTypeBytes:
		return "Bytes"
	case ControlTypeIEC958:
		return "IEC958"
	case ControlTypeInteger64:
		return "Integer64"
	default:
		return "None"
	}
}

// PortCategory represents the routing port category
type PortCategory int

const (
	PortCategoryOff PortCategory = iota
	PortCategoryHW
	PortCategoryMix
	PortCategoryDSP
	PortCategoryPCM
)

func (p PortCategory) String() string {
	switch p {
	case PortCategoryOff:
		return "Off"
	case PortCategoryHW:
		return "Hardware"
	case PortCategoryMix:
		return "Mixer"
	case PortCategoryDSP:
		return "DSP"
	case PortCategoryPCM:
		return "PCM"
	default:
		return "Unknown"
	}
}

// Card represents a Scarlett audio interface card
type Card struct {
	Number int
	Name   string
	handle *alsaHandle
}

// Control represents an ALSA control element
type Control struct {
	NumID   uint
	Name    string
	Type    ControlType
	Count   int
	Index   int
	card    *Card
	// for integer/enumerated types
	Min int64
	Max int64
	// for enumerated types
	Items []string
}

// RoutingSource represents a routing source endpoint
type RoutingSource struct {
	ID           int
	Name         string
	Category     PortCategory
	PortNum      int
	HardwareType string
}

// RoutingSink represents a routing sink (destination)
type RoutingSink struct {
	Index    int
	Name     string
	Category PortCategory
	PortNum  int
	Control  *Control
}

// EventCallback is called when a control changes value
type EventCallback func(control *Control)

// alsaHandle wraps the C ALSA control handle (internal use only)
type alsaHandle struct {
	ptr     uintptr // snd_ctl_t* as uintptr
	pollFds []int
}
