package scarlettctl

import (
	"fmt"
	"strings"
)

// GetControls returns all controls for this card
func (c *Card) GetControls() ([]*Control, error) {
	if c.handle == nil {
		return nil, fmt.Errorf("card not open")
	}

	controls, err := enumerateControls(c.handle)
	if err != nil {
		return nil, err
	}

	// link controls back to their card
	for _, ctl := range controls {
		ctl.card = c
	}

	return controls, nil
}

// FindControl finds a control by exact name or full ID
// If the input contains ':' and '/', it is treated as a full ID (e.g., "mixer:0.0/Level Meter[0]")
// Otherwise it is treated as a control name
func (c *Card) FindControl(name string) (*Control, error) {
	// try full ID lookup if input looks like an ID
	if strings.Contains(name, ":") && strings.Contains(name, "/") {
		return c.FindControlByID(name)
	}

	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	for _, ctl := range controls {
		if ctl.Name == name {
			return ctl, nil
		}
	}

	return nil, fmt.Errorf("control '%s' not found", name)
}

// FindControlByID finds a control by its full identifier
// The ID format is "interface:device.subdevice/name[index]" (e.g., "mixer:0.0/Level Meter[0]")
func (c *Card) FindControlByID(id string) (*Control, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	for _, ctl := range controls {
		if ctl.FullID() == id {
			return ctl, nil
		}
	}

	return nil, fmt.Errorf("control with id '%s' not found", id)
}

// FindControlByPrefix finds a control by name prefix
func (c *Card) FindControlByPrefix(prefix string) (*Control, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	for _, ctl := range controls {
		if strings.HasPrefix(ctl.Name, prefix) {
			return ctl, nil
		}
	}

	return nil, fmt.Errorf("control with prefix '%s' not found", prefix)
}

// FindControlsMatching finds all controls matching a pattern
func (c *Card) FindControlsMatching(pattern string) ([]*Control, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	var matched []*Control
	patternLower := strings.ToLower(pattern)

	for _, ctl := range controls {
		if strings.Contains(strings.ToLower(ctl.Name), patternLower) {
			matched = append(matched, ctl)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no controls matching '%s' found", pattern)
	}

	return matched, nil
}

// GetValue reads the current value of the control
func (ctl *Control) GetValue() (int64, error) {
	if ctl.card == nil || ctl.card.handle == nil {
		return 0, fmt.Errorf("control not associated with open card")
	}

	return readControl(ctl.card.handle, ctl)
}

// SetValue writes a value to the control
func (ctl *Control) SetValue(value int64) error {
	if ctl.card == nil || ctl.card.handle == nil {
		return fmt.Errorf("control not associated with open card")
	}

	// validate value range for integer types
	if ctl.Type == ControlTypeInteger || ctl.Type == ControlTypeInteger64 {
		if value < ctl.Min || value > ctl.Max {
			return fmt.Errorf("value %d out of range [%d, %d]", value, ctl.Min, ctl.Max)
		}
	}

	// validate enum index
	if ctl.Type == ControlTypeEnumerated {
		if value < 0 || value >= int64(len(ctl.Items)) {
			return fmt.Errorf("enum index %d out of range [0, %d]", value, len(ctl.Items)-1)
		}
	}

	return writeControl(ctl.card.handle, ctl, value)
}

// GetValueString returns the control value as a human-readable string
func (ctl *Control) GetValueString() (string, error) {
	value, err := ctl.GetValue()
	if err != nil {
		return "", err
	}

	switch ctl.Type {
	case ControlTypeBoolean:
		if value == 0 {
			return "Off", nil
		}
		return "On", nil

	case ControlTypeEnumerated:
		if value >= 0 && value < int64(len(ctl.Items)) {
			return ctl.Items[value], nil
		}
		return fmt.Sprintf("Unknown(%d)", value), nil

	case ControlTypeInteger, ControlTypeInteger64:
		return fmt.Sprintf("%d", value), nil

	default:
		return fmt.Sprintf("%d", value), nil
	}
}

// SetValueByString sets the control value from a string representation
func (ctl *Control) SetValueByString(valueStr string) error {
	switch ctl.Type {
	case ControlTypeBoolean:
		lowerVal := strings.ToLower(valueStr)
		if lowerVal == "on" || lowerVal == "true" || lowerVal == "1" || lowerVal == "yes" {
			return ctl.SetValue(1)
		}
		if lowerVal == "off" || lowerVal == "false" || lowerVal == "0" || lowerVal == "no" {
			return ctl.SetValue(0)
		}
		return fmt.Errorf("invalid boolean value: %s (use on/off, true/false, 1/0, yes/no)", valueStr)

	case ControlTypeEnumerated:
		// try to find matching enum item
		for i, item := range ctl.Items {
			if strings.EqualFold(item, valueStr) {
				return ctl.SetValue(int64(i))
			}
		}
		// try parsing as index
		var index int64
		if _, err := fmt.Sscanf(valueStr, "%d", &index); err == nil {
			return ctl.SetValue(index)
		}
		return fmt.Errorf("invalid enum value: %s (valid: %v)", valueStr, ctl.Items)

	case ControlTypeInteger, ControlTypeInteger64:
		var value int64
		if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
			return fmt.Errorf("invalid integer value: %s", valueStr)
		}
		return ctl.SetValue(value)

	default:
		return fmt.Errorf("unsupported control type: %v", ctl.Type)
	}
}

// String returns a string representation of the control
func (ctl *Control) String() string {
	var sb strings.Builder

	// show interface/device/subdevice prefix for disambiguation
	sb.WriteString(fmt.Sprintf("[%s:%d.%d] ", ctl.Interface, ctl.Device, ctl.Subdevice))
	sb.WriteString(fmt.Sprintf("%-50s [%s]", ctl.Name, ctl.Type))

	switch ctl.Type {
	case ControlTypeInteger, ControlTypeInteger64:
		sb.WriteString(fmt.Sprintf(" range: [%d, %d]", ctl.Min, ctl.Max))
	case ControlTypeEnumerated:
		sb.WriteString(fmt.Sprintf(" items: %v", ctl.Items))
	}

	if ctl.Count > 1 {
		sb.WriteString(fmt.Sprintf(" (index %d of %d)", ctl.Index, ctl.Count))
	}

	return sb.String()
}

// FullID returns a unique identifier string for the control
func (ctl *Control) FullID() string {
	return fmt.Sprintf("%s:%d.%d/%s[%d]", ctl.Interface, ctl.Device, ctl.Subdevice, ctl.Name, ctl.Index)
}

// DetailedString returns a detailed string representation including current value
func (ctl *Control) DetailedString() string {
	value, err := ctl.GetValueString()
	if err != nil {
		value = fmt.Sprintf("Error: %v", err)
	}

	return fmt.Sprintf("%s = %s", ctl.String(), value)
}
