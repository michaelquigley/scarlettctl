package scarlettctl

import (
	"fmt"
	"regexp"
)

// MixerInput represents a mixer input channel
type MixerInput struct {
	MixName   string // e.g., "Mix A", "Mix B"
	InputNum  int    // 1-based input number
	Control   *Control
}

// GetMixerInputs returns all mixer input volume controls
func (c *Card) GetMixerInputs() ([]MixerInput, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	var inputs []MixerInput

	// gen 2/3/4 pattern: "Mix A Input 01 Playback Volume"
	// gen 1 pattern: "Matrix 01 Mix A Playback Volume"
	gen234Re := regexp.MustCompile(`^Mix ([A-Z]) Input (\d+) Playback Volume$`)
	gen1Re := regexp.MustCompile(`^Matrix (\d+) Mix ([A-Z]) Playback Volume$`)

	for _, ctl := range controls {
		if ctl.Type != ControlTypeInteger {
			continue
		}

		// try gen 2/3/4 pattern
		if matches := gen234Re.FindStringSubmatch(ctl.Name); matches != nil {
			mixName := "Mix " + matches[1]
			var inputNum int
			fmt.Sscanf(matches[2], "%d", &inputNum)

			inputs = append(inputs, MixerInput{
				MixName:  mixName,
				InputNum: inputNum,
				Control:  ctl,
			})
			continue
		}

		// try gen 1 pattern
		if matches := gen1Re.FindStringSubmatch(ctl.Name); matches != nil {
			var inputNum int
			fmt.Sscanf(matches[1], "%d", &inputNum)
			mixName := "Mix " + matches[2]

			inputs = append(inputs, MixerInput{
				MixName:  mixName,
				InputNum: inputNum,
				Control:  ctl,
			})
		}
	}

	return inputs, nil
}

// GetMixerInput gets a specific mixer input control
func (c *Card) GetMixerInput(mixName string, inputNum int) (*Control, error) {
	inputs, err := c.GetMixerInputs()
	if err != nil {
		return nil, err
	}

	for _, input := range inputs {
		if input.MixName == mixName && input.InputNum == inputNum {
			return input.Control, nil
		}
	}

	return nil, fmt.Errorf("mixer input %s #%d not found", mixName, inputNum)
}

// SetMixerLevel sets a mixer input level
func (c *Card) SetMixerLevel(mixName string, inputNum int, level int64) error {
	ctl, err := c.GetMixerInput(mixName, inputNum)
	if err != nil {
		return err
	}

	return ctl.SetValue(level)
}

// GetMixerLevel gets a mixer input level
func (c *Card) GetMixerLevel(mixName string, inputNum int) (int64, error) {
	ctl, err := c.GetMixerInput(mixName, inputNum)
	if err != nil {
		return 0, err
	}

	return ctl.GetValue()
}

// PrintMixerState prints the current state of all mixer inputs
func (c *Card) PrintMixerState() error {
	inputs, err := c.GetMixerInputs()
	if err != nil {
		return err
	}

	if len(inputs) == 0 {
		fmt.Println("no mixer controls found")
		return nil
	}

	fmt.Println("\nmixer state:")
	fmt.Println("============")

	currentMix := ""
	for _, input := range inputs {
		if input.MixName != currentMix {
			if currentMix != "" {
				fmt.Println()
			}
			fmt.Printf("%s:\n", input.MixName)
			currentMix = input.MixName
		}

		value, err := input.Control.GetValue()
		if err != nil {
			fmt.Printf("  input %02d: error - %v\n", input.InputNum, err)
			continue
		}

		// show value and range
		fmt.Printf("  input %02d: %5d [%d..%d]\n",
			input.InputNum, value, input.Control.Min, input.Control.Max)
	}

	return nil
}
