package scarlettctl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// routing sink patterns (gen 2/3/4)
	routingSinkPatterns = []string{
		"PCM %02d Capture Enum",
		"Input Source %02d Capture Route",
		"Mixer Input %02d Capture Enum",
		"Matrix %02d Capture Enum",
		"DSP Input %02d Capture Enum",
		"Analogue Output %02d Playback Enum",
		"S/PDIF Output %02d Playback Enum",
		"ADAT Output %02d Playback Enum",
	}

	// port category detection regexes
	portCategoryRegexes = map[PortCategory]*regexp.Regexp{
		PortCategoryPCM: regexp.MustCompile(`^PCM \d+`),
		PortCategoryMix: regexp.MustCompile(`^(Mixer Input|Mixer|Matrix) \d+`),
		PortCategoryDSP: regexp.MustCompile(`^DSP Input \d+`),
		PortCategoryHW:  regexp.MustCompile(`^(Analogue|S/PDIF|ADAT)( Output| Input)? \d+`),
	}
)

// GetRoutingSources returns all routing sources available on the card
func (c *Card) GetRoutingSources() ([]RoutingSource, error) {
	// find a routing sink control to extract source names from
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	var sinkControl *Control
	for _, ctl := range controls {
		if ctl.Type == ControlTypeEnumerated && isRoutingSink(ctl.Name) {
			sinkControl = ctl
			break
		}
	}

	if sinkControl == nil {
		return nil, fmt.Errorf("no routing controls found")
	}

	sources := make([]RoutingSource, 0, len(sinkControl.Items))

	for i, name := range sinkControl.Items {
		category, portNum := parseRoutingSourceName(name)

		src := RoutingSource{
			ID:       i,
			Name:     name,
			Category: category,
			PortNum:  portNum,
		}

		// detect hardware type from name
		if strings.Contains(name, "Analogue") {
			src.HardwareType = "Analogue"
		} else if strings.Contains(name, "S/PDIF") {
			src.HardwareType = "S/PDIF"
		} else if strings.Contains(name, "ADAT") {
			src.HardwareType = "ADAT"
		}

		sources = append(sources, src)
	}

	return sources, nil
}

// GetRoutingSinks returns all routing sinks (destinations) on the card
func (c *Card) GetRoutingSinks() ([]RoutingSink, error) {
	controls, err := c.GetControls()
	if err != nil {
		return nil, err
	}

	sinks := make([]RoutingSink, 0)
	sinkIndex := 0

	for _, ctl := range controls {
		if ctl.Type == ControlTypeEnumerated && isRoutingSink(ctl.Name) {
			category, portNum := parseRoutingSinkName(ctl.Name)

			sink := RoutingSink{
				Index:    sinkIndex,
				Name:     ctl.Name,
				Category: category,
				PortNum:  portNum,
				Control:  ctl,
			}

			sinks = append(sinks, sink)
			sinkIndex++
		}
	}

	if len(sinks) == 0 {
		return nil, fmt.Errorf("no routing sinks found")
	}

	return sinks, nil
}

// GetRouting returns the current routing configuration as a map of sink -> source ID
func (c *Card) GetRouting() (map[string]int, error) {
	sinks, err := c.GetRoutingSinks()
	if err != nil {
		return nil, err
	}

	routing := make(map[string]int)

	for _, sink := range sinks {
		value, err := sink.Control.GetValue()
		if err != nil {
			return nil, fmt.Errorf("failed to read routing for %s: %v", sink.Name, err)
		}
		routing[sink.Name] = int(value)
	}

	return routing, nil
}

// SetRouting sets a routing connection
func (c *Card) SetRouting(sinkName string, sourceID int) error {
	sinks, err := c.GetRoutingSinks()
	if err != nil {
		return err
	}

	for _, sink := range sinks {
		if sink.Name == sinkName {
			return sink.Control.SetValue(int64(sourceID))
		}
	}

	return fmt.Errorf("routing sink '%s' not found", sinkName)
}

// SetRoutingByNames sets a routing connection using source and sink names
func (c *Card) SetRoutingByNames(sinkName, sourceName string) error {
	// find the sink
	sinks, err := c.GetRoutingSinks()
	if err != nil {
		return err
	}

	var targetSink *RoutingSink
	for i := range sinks {
		if strings.Contains(sinks[i].Name, sinkName) {
			targetSink = &sinks[i]
			break
		}
	}

	if targetSink == nil {
		return fmt.Errorf("routing sink matching '%s' not found", sinkName)
	}

	// find the source ID
	sources, err := c.GetRoutingSources()
	if err != nil {
		return err
	}

	for _, src := range sources {
		if strings.Contains(src.Name, sourceName) || src.Name == sourceName {
			return targetSink.Control.SetValue(int64(src.ID))
		}
	}

	return fmt.Errorf("routing source matching '%s' not found", sourceName)
}

// isRoutingSink checks if a control name matches routing sink patterns
func isRoutingSink(name string) bool {
	// check for "Capture Enum" or "Playback Enum" which are routing controls
	return (strings.Contains(name, "Capture Enum") ||
	        strings.Contains(name, "Playback Enum") ||
	        strings.Contains(name, "Capture Route")) &&
	       !strings.Contains(name, "Volume") &&
	       !strings.Contains(name, "Switch")
}

// parseRoutingSinkName extracts category and port number from sink name
func parseRoutingSinkName(name string) (PortCategory, int) {
	for category, regex := range portCategoryRegexes {
		if regex.MatchString(name) {
			portNum := extractPortNumber(name)
			return category, portNum
		}
	}
	return PortCategoryOff, 0
}

// parseRoutingSourceName extracts category and port number from source name
func parseRoutingSourceName(name string) (PortCategory, int) {
	if name == "Off" {
		return PortCategoryOff, 0
	}

	// check for Mix A, Mix B, etc.
	if strings.HasPrefix(name, "Mix ") {
		letter := name[4:5]
		portNum := int(letter[0] - 'A')
		return PortCategoryMix, portNum
	}

	// check for PCM XX
	if strings.HasPrefix(name, "PCM ") {
		portNum := extractPortNumber(name)
		return PortCategoryPCM, portNum - 1 // PCM is 1-indexed in names
	}

	// check for DSP X
	if strings.HasPrefix(name, "DSP ") {
		portNum := extractPortNumber(name)
		return PortCategoryDSP, portNum - 1
	}

	// check for hardware (Analogue, S/PDIF, ADAT)
	if strings.Contains(name, "Analogue") ||
	   strings.Contains(name, "S/PDIF") ||
	   strings.Contains(name, "ADAT") {
		portNum := extractPortNumber(name)
		return PortCategoryHW, portNum - 1
	}

	return PortCategoryOff, 0
}

// extractPortNumber extracts a number from a string
func extractPortNumber(s string) int {
	// find all numbers in the string
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(s, -1)

	if len(matches) > 0 {
		num, err := strconv.Atoi(matches[0])
		if err == nil {
			return num
		}
	}

	return 0
}

// PrintRoutingMatrix prints a human-readable routing matrix
func (c *Card) PrintRoutingMatrix() error {
	sources, err := c.GetRoutingSources()
	if err != nil {
		return err
	}

	sinks, err := c.GetRoutingSinks()
	if err != nil {
		return err
	}

	// print available sources organized by category
	fmt.Println("\n════════════════════════════════════════════════════════════")
	fmt.Println("                    routing sources")
	fmt.Println("════════════════════════════════════════════════════════════")

	printSourcesByCategory := func(category PortCategory, title string) {
		var categorySource []RoutingSource
		for _, src := range sources {
			if src.Category == category {
				categorySource = append(categorySource, src)
			}
		}

		if len(categorySource) > 0 {
			fmt.Printf("\n%s:\n", title)
			for _, src := range categorySource {
				hwType := ""
				if src.HardwareType != "" {
					hwType = fmt.Sprintf(" [%s]", src.HardwareType)
				}
				fmt.Printf("  [%2d] %-20s %s%s\n", src.ID, src.Name, src.Category, hwType)
			}
		}
	}

	printSourcesByCategory(PortCategoryOff, "off")
	printSourcesByCategory(PortCategoryHW, "hardware inputs")
	printSourcesByCategory(PortCategoryMix, "mixer outputs")
	printSourcesByCategory(PortCategoryPCM, "PCM (computer playback)")
	printSourcesByCategory(PortCategoryDSP, "dsp outputs")

	// print routing organized by sink category
	fmt.Println("\n════════════════════════════════════════════════════════════")
	fmt.Println("                    routing matrix")
	fmt.Println("════════════════════════════════════════════════════════════")

	printSinksByCategory := func(category PortCategory, title string) {
		var categorySinks []RoutingSink
		for _, sink := range sinks {
			if sink.Category == category {
				categorySinks = append(categorySinks, sink)
			}
		}

		if len(categorySinks) > 0 {
			fmt.Printf("\n%s:\n", title)
			fmt.Println(strings.Repeat("-", 60))

			for _, sink := range categorySinks {
				value, err := sink.Control.GetValue()
				if err != nil {
					fmt.Printf("  %-35s -> error: %v\n", sink.Name, err)
					continue
				}

				sourceName := "unknown"
				sourceInfo := ""
				if value >= 0 && int(value) < len(sources) {
					src := sources[value]
					sourceName = src.Name
					if src.Category != PortCategoryOff {
						sourceInfo = fmt.Sprintf(" (%s)", src.Category)
						if src.HardwareType != "" {
							sourceInfo = fmt.Sprintf(" (%s %s)", src.Category, src.HardwareType)
						}
					}
				}

				fmt.Printf("  %-35s <- %-20s%s\n",
					shortSinkName(sink.Name),
					sourceName,
					sourceInfo)
			}
		}
	}

	printSinksByCategory(PortCategoryHW, "hardware outputs (to speakers/monitors)")
	printSinksByCategory(PortCategoryPCM, "PCM capture (to computer/DAW)")
	printSinksByCategory(PortCategoryMix, "mixer inputs")
	printSinksByCategory(PortCategoryDSP, "dsp inputs")

	fmt.Println("\n════════════════════════════════════════════════════════════")
	fmt.Printf("total: %d sources, %d sinks\n", len(sources), len(sinks))
	fmt.Println("════════════════════════════════════════════════════════════\n")

	return nil
}

// shortSinkName shortens sink control names for display
func shortSinkName(name string) string {
	// remove redundant suffixes
	name = strings.TrimSuffix(name, " Playback Enum")
	name = strings.TrimSuffix(name, " Capture Enum")
	name = strings.TrimSuffix(name, " Capture Route")
	return name
}
