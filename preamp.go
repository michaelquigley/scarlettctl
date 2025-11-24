package scarlettctl

import (
	"fmt"
	"regexp"
)

// PreampChannel represents a preamp input channel with all its controls
type PreampChannel struct {
	ChannelNum    int
	Gain          *Control
	Phantom       *Control
	Air           *Control
	Pad           *Control
	Impedance     *Control
	Level         *Control
	Autogain      *Control
	Safe          *Control
	Link          *Control
}

// GetPreampChannels returns all preamp channels with their controls
func (c *Card) GetPreampChannels() ([]PreampChannel, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	// build a map of channel number -> controls
	channelMap := make(map[int]*PreampChannel)

	gainRe := regexp.MustCompile(`^Line In (\d+) Gain Capture Volume$`)
	phantomRe := regexp.MustCompile(`^Line In (\d+)(?:-\d+)? Phantom Power Capture Switch$`)
	airRe := regexp.MustCompile(`^Line In (\d+) Air Capture (?:Switch|Enum)$`)
	padRe := regexp.MustCompile(`^Line In (\d+) Pad Capture Switch$`)
	impedanceRe := regexp.MustCompile(`^Line In (\d+) Impedance Switch$`)
	levelRe := regexp.MustCompile(`^Line In (\d+) Level Capture Enum$`)
	autogainRe := regexp.MustCompile(`^Line In (\d+) Autogain Capture Switch$`)
	safeRe := regexp.MustCompile(`^Line In (\d+) Safe Capture Switch$`)
	linkRe := regexp.MustCompile(`^Line In (\d+)-\d+ Link Capture Switch$`)

	for _, ctl := range controls {
		var channelNum int

		// match against each pattern
		if matches := gainRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Gain = ctl
		} else if matches := phantomRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Phantom = ctl
		} else if matches := airRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Air = ctl
		} else if matches := padRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Pad = ctl
		} else if matches := impedanceRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Impedance = ctl
		} else if matches := levelRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Level = ctl
		} else if matches := autogainRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Autogain = ctl
		} else if matches := safeRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Safe = ctl
		} else if matches := linkRe.FindStringSubmatch(ctl.Name); matches != nil {
			fmt.Sscanf(matches[1], "%d", &channelNum)
			if _, exists := channelMap[channelNum]; !exists {
				channelMap[channelNum] = &PreampChannel{ChannelNum: channelNum}
			}
			channelMap[channelNum].Link = ctl
		}
	}

	// convert map to sorted slice
	channels := make([]PreampChannel, 0, len(channelMap))
	for i := 1; i <= len(channelMap)+10; i++ { // +10 to handle gaps
		if ch, exists := channelMap[i]; exists {
			channels = append(channels, *ch)
		}
	}

	return channels, nil
}

// GetPreampChannel gets a specific preamp channel
func (c *Card) GetPreampChannel(channelNum int) (*PreampChannel, error) {
	channels, err := c.GetPreampChannels()
	if err != nil {
		return nil, err
	}

	for i := range channels {
		if channels[i].ChannelNum == channelNum {
			return &channels[i], nil
		}
	}

	return nil, fmt.Errorf("preamp channel %d not found", channelNum)
}

// SetPreampGain sets the gain for a preamp channel
func (c *Card) SetPreampGain(channelNum int, gain int64) error {
	ch, err := c.GetPreampChannel(channelNum)
	if err != nil {
		return err
	}

	if ch.Gain == nil {
		return fmt.Errorf("channel %d has no gain control", channelNum)
	}

	return ch.Gain.SetValue(gain)
}

// SetPreampPhantom sets phantom power for a preamp channel
func (c *Card) SetPreampPhantom(channelNum int, enabled bool) error {
	ch, err := c.GetPreampChannel(channelNum)
	if err != nil {
		return err
	}

	if ch.Phantom == nil {
		return fmt.Errorf("channel %d has no phantom power control", channelNum)
	}

	value := int64(0)
	if enabled {
		value = 1
	}

	return ch.Phantom.SetValue(value)
}

// SetPreampAir sets air mode for a preamp channel
func (c *Card) SetPreampAir(channelNum int, enabled bool) error {
	ch, err := c.GetPreampChannel(channelNum)
	if err != nil {
		return err
	}

	if ch.Air == nil {
		return fmt.Errorf("channel %d has no air control", channelNum)
	}

	value := int64(0)
	if enabled {
		value = 1
	}

	return ch.Air.SetValue(value)
}

// SetPreampPad sets pad for a preamp channel
func (c *Card) SetPreampPad(channelNum int, enabled bool) error {
	ch, err := c.GetPreampChannel(channelNum)
	if err != nil {
		return err
	}

	if ch.Pad == nil {
		return fmt.Errorf("channel %d has no pad control", channelNum)
	}

	value := int64(0)
	if enabled {
		value = 1
	}

	return ch.Pad.SetValue(value)
}

// PrintPreampState prints the current state of all preamp channels
func (c *Card) PrintPreampState() error {
	channels, err := c.GetPreampChannels()
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		fmt.Println("no preamp controls found")
		return nil
	}

	fmt.Println("\npreamp state:")
	fmt.Println("=============")

	for _, ch := range channels {
		fmt.Printf("\nchannel %d:\n", ch.ChannelNum)

		if ch.Gain != nil {
			value, _ := ch.Gain.GetValueString()
			fmt.Printf("  gain:         %s [%d..%d]\n", value, ch.Gain.Min, ch.Gain.Max)
		}

		if ch.Phantom != nil {
			value, _ := ch.Phantom.GetValueString()
			fmt.Printf("  phantom 48v:  %s\n", value)
		}

		if ch.Air != nil {
			value, _ := ch.Air.GetValueString()
			fmt.Printf("  air:          %s\n", value)
		}

		if ch.Pad != nil {
			value, _ := ch.Pad.GetValueString()
			fmt.Printf("  pad:          %s\n", value)
		}

		if ch.Impedance != nil {
			value, _ := ch.Impedance.GetValueString()
			fmt.Printf("  impedance:    %s\n", value)
		}

		if ch.Level != nil {
			value, _ := ch.Level.GetValueString()
			fmt.Printf("  level:        %s\n", value)
		}

		if ch.Autogain != nil {
			value, _ := ch.Autogain.GetValueString()
			fmt.Printf("  autogain:     %s\n", value)
		}

		if ch.Safe != nil {
			value, _ := ch.Safe.GetValueString()
			fmt.Printf("  safe:         %s\n", value)
		}

		if ch.Link != nil {
			value, _ := ch.Link.GetValueString()
			fmt.Printf("  link:         %s\n", value)
		}
	}

	return nil
}
