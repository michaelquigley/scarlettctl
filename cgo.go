package scarlettctl

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <stdlib.h>
#include <string.h>

// Helper to get enum item name
static int get_enum_item_name(snd_ctl_t *handle, snd_ctl_elem_info_t *info, unsigned int idx, char *buf, size_t size) {
	snd_ctl_elem_info_set_item(info, idx);
	// Must call snd_ctl_elem_info again after setting item to query that item's name
	int err = snd_ctl_elem_info(handle, info);
	if (err < 0) {
		return err;
	}
	const char *name = snd_ctl_elem_info_get_item_name(info);
	if (!name) {
		return -1;
	}
	strncpy(buf, name, size - 1);
	buf[size - 1] = '\0';
	return 0;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// alsaError converts ALSA error codes to Go errors
func alsaError(code C.int, operation string) error {
	if code >= 0 {
		return nil
	}
	errStr := C.GoString(C.snd_strerror(code))
	return fmt.Errorf("%s: %s", operation, errStr)
}

// openCard opens an ALSA control handle for the specified card number
func openCard(cardNum int) (*alsaHandle, error) {
	var handle *C.snd_ctl_t
	cardName := fmt.Sprintf("hw:%d", cardNum)
	cCardName := C.CString(cardName)
	defer C.free(unsafe.Pointer(cCardName))

	err := C.snd_ctl_open(&handle, cCardName, 0)
	if err < 0 {
		return nil, alsaError(err, "open card")
	}

	// subscribe to events
	err = C.snd_ctl_subscribe_events(handle, 1)
	if err < 0 {
		C.snd_ctl_close(handle)
		return nil, alsaError(err, "subscribe to events")
	}

	// get poll descriptors
	count := C.snd_ctl_poll_descriptors_count(handle)
	if count <= 0 {
		C.snd_ctl_close(handle)
		return nil, fmt.Errorf("no poll descriptors available")
	}

	pfds := make([]C.struct_pollfd, count)
	n := C.snd_ctl_poll_descriptors(handle, &pfds[0], C.uint(count))
	if n < 0 {
		C.snd_ctl_close(handle)
		return nil, alsaError(n, "get poll descriptors")
	}

	pollFds := make([]int, count)
	for i := 0; i < int(count); i++ {
		pollFds[i] = int(pfds[i].fd)
	}

	return &alsaHandle{
		ptr:     uintptr(unsafe.Pointer(handle)),
		pollFds: pollFds,
	}, nil
}

// closeCard closes an ALSA control handle
func closeCard(h *alsaHandle) error {
	if h == nil || h.ptr == 0 {
		return nil
	}
	handle := (*C.snd_ctl_t)(unsafe.Pointer(h.ptr))
	err := C.snd_ctl_close(handle)
	h.ptr = 0
	return alsaError(err, "close card")
}

// getCardInfo retrieves card information
func getCardInfo(cardNum int) (string, error) {
	var info *C.snd_ctl_card_info_t
	C.snd_ctl_card_info_malloc(&info)
	defer C.snd_ctl_card_info_free(info)

	var handle *C.snd_ctl_t
	cardName := fmt.Sprintf("hw:%d", cardNum)
	cCardName := C.CString(cardName)
	defer C.free(unsafe.Pointer(cCardName))

	err := C.snd_ctl_open(&handle, cCardName, 0)
	if err < 0 {
		return "", alsaError(err, "open card for info")
	}
	defer C.snd_ctl_close(handle)

	err = C.snd_ctl_card_info(handle, info)
	if err < 0 {
		return "", alsaError(err, "get card info")
	}

	name := C.GoString(C.snd_ctl_card_info_get_name(info))
	return name, nil
}

// enumerateControls lists all controls on a card
func enumerateControls(h *alsaHandle) ([]*Control, error) {
	handle := (*C.snd_ctl_t)(unsafe.Pointer(h.ptr))
	var info *C.snd_ctl_elem_info_t
	C.snd_ctl_elem_info_malloc(&info)
	defer C.snd_ctl_elem_info_free(info)

	var list *C.snd_ctl_elem_list_t
	C.snd_ctl_elem_list_malloc(&list)
	defer C.snd_ctl_elem_list_free(list)

	err := C.snd_ctl_elem_list(handle, list)
	if err < 0 {
		return nil, alsaError(err, "get element list")
	}

	count := C.snd_ctl_elem_list_get_count(list)
	err = C.snd_ctl_elem_list_alloc_space(list, count)
	if err < 0 {
		return nil, alsaError(err, "allocate element list space")
	}
	defer C.snd_ctl_elem_list_free_space(list)

	err = C.snd_ctl_elem_list(handle, list)
	if err < 0 {
		return nil, alsaError(err, "fill element list")
	}

	controls := make([]*Control, 0, count)

	for i := C.uint(0); i < count; i++ {
		numid := C.snd_ctl_elem_list_get_numid(list, i)

		C.snd_ctl_elem_info_set_numid(info, numid)
		err = C.snd_ctl_elem_info(handle, info)
		if err < 0 {
			continue // skip controls we can't query
		}

		// get basic info
		name := C.GoString(C.snd_ctl_elem_info_get_name(info))
		ctlType := ControlType(C.snd_ctl_elem_info_get_type(info))
		ctlCount := int(C.snd_ctl_elem_info_get_count(info))

		// get element identifier metadata
		ctlInterface := InterfaceType(C.snd_ctl_elem_info_get_interface(info))
		ctlDevice := uint(C.snd_ctl_elem_info_get_device(info))
		ctlSubdevice := uint(C.snd_ctl_elem_info_get_subdevice(info))

		// create control for each value in multi-value controls
		for idx := 0; idx < ctlCount; idx++ {
			ctl := &Control{
				NumID:     uint(numid),
				Name:      name,
				Type:      ctlType,
				Count:     ctlCount,
				Index:     idx,
				Interface: ctlInterface,
				Device:    ctlDevice,
				Subdevice: ctlSubdevice,
			}

			// get type-specific information
			switch ctlType {
			case ControlTypeInteger:
				ctl.Min = int64(C.snd_ctl_elem_info_get_min(info))
				ctl.Max = int64(C.snd_ctl_elem_info_get_max(info))

			case ControlTypeInteger64:
				ctl.Min = int64(C.snd_ctl_elem_info_get_min64(info))
				ctl.Max = int64(C.snd_ctl_elem_info_get_max64(info))

			case ControlTypeEnumerated:
				itemCount := C.snd_ctl_elem_info_get_items(info)
				ctl.Items = make([]string, itemCount)

				buf := make([]byte, 256)
				for j := C.uint(0); j < itemCount; j++ {
					if C.get_enum_item_name(handle, info, j, (*C.char)(unsafe.Pointer(&buf[0])), 256) == 0 {
						ctl.Items[j] = string(buf[:cstrlen(buf)])
					}
				}
			}

			controls = append(controls, ctl)
		}
	}

	return controls, nil
}

// readControl reads the current value of a control
func readControl(h *alsaHandle, ctl *Control) (int64, error) {
	handle := (*C.snd_ctl_t)(unsafe.Pointer(h.ptr))
	var value *C.snd_ctl_elem_value_t
	C.snd_ctl_elem_value_malloc(&value)
	defer C.snd_ctl_elem_value_free(value)

	C.snd_ctl_elem_value_set_numid(value, C.uint(ctl.NumID))
	err := C.snd_ctl_elem_read(handle, value)
	if err < 0 {
		return 0, alsaError(err, "read control")
	}

	var result C.long
	switch ctl.Type {
	case ControlTypeBoolean:
		result = C.long(C.snd_ctl_elem_value_get_boolean(value, C.uint(ctl.Index)))
	case ControlTypeInteger:
		result = C.snd_ctl_elem_value_get_integer(value, C.uint(ctl.Index))
	case ControlTypeEnumerated:
		result = C.long(C.snd_ctl_elem_value_get_enumerated(value, C.uint(ctl.Index)))
	case ControlTypeInteger64:
		return int64(C.snd_ctl_elem_value_get_integer64(value, C.uint(ctl.Index))), nil
	default:
		return 0, fmt.Errorf("unsupported control type: %v", ctl.Type)
	}

	return int64(result), nil
}

// writeControl writes a value to a control
func writeControl(h *alsaHandle, ctl *Control, value int64) error {
	handle := (*C.snd_ctl_t)(unsafe.Pointer(h.ptr))
	var elemValue *C.snd_ctl_elem_value_t
	C.snd_ctl_elem_value_malloc(&elemValue)
	defer C.snd_ctl_elem_value_free(elemValue)

	// read current value first
	C.snd_ctl_elem_value_set_numid(elemValue, C.uint(ctl.NumID))
	err := C.snd_ctl_elem_read(handle, elemValue)
	if err < 0 {
		return alsaError(err, "read before write")
	}

	// set the new value
	switch ctl.Type {
	case ControlTypeBoolean:
		C.snd_ctl_elem_value_set_boolean(elemValue, C.uint(ctl.Index), C.long(value))
	case ControlTypeInteger:
		C.snd_ctl_elem_value_set_integer(elemValue, C.uint(ctl.Index), C.long(value))
	case ControlTypeEnumerated:
		C.snd_ctl_elem_value_set_enumerated(elemValue, C.uint(ctl.Index), C.uint(value))
	case ControlTypeInteger64:
		C.snd_ctl_elem_value_set_integer64(elemValue, C.uint(ctl.Index), C.longlong(value))
	default:
		return fmt.Errorf("unsupported control type for writing: %v", ctl.Type)
	}

	// write it back
	err = C.snd_ctl_elem_write(handle, elemValue)
	return alsaError(err, "write control")
}

// checkEvent checks if there's a pending event
func checkEvent(h *alsaHandle) (bool, error) {
	handle := (*C.snd_ctl_t)(unsafe.Pointer(h.ptr))
	var event *C.snd_ctl_event_t
	C.snd_ctl_event_malloc(&event)
	defer C.snd_ctl_event_free(event)

	err := C.snd_ctl_read(handle, event)
	if err < 0 {
		if err == -C.EAGAIN {
			return false, nil // no event available
		}
		return false, alsaError(err, "read event")
	}

	// check if it's an element event
	eventType := C.snd_ctl_event_get_type(event)
	if eventType == C.SND_CTL_EVENT_ELEM {
		return true, nil
	}

	return false, nil
}

// listCardNumbers returns the indices of all available ALSA cards
func listCardNumbers() ([]int, error) {
	var cardNum C.int = -1
	var cards []int

	for {
		err := C.snd_card_next(&cardNum)
		if err < 0 {
			return nil, alsaError(err, "enumerate cards")
		}
		if cardNum < 0 {
			break // no more cards
		}
		cards = append(cards, int(cardNum))
	}

	return cards, nil
}

// cstrlen finds the length of a null-terminated C string in a byte slice
func cstrlen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}
