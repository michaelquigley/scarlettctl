package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/michaelquigley/scarlettctl"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "scarlettctl",
	Short: "Control Focusrite Scarlett audio interfaces",
	Long: `scarlettctl is a command-line tool for controlling Focusrite Scarlett,
Vocaster, and Clarett audio interfaces via the ALSA control interface.

It provides access to mixer controls, routing, preamp settings, and more.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available Scarlett devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		cards, err := scarlettctl.ListCards()
		if err != nil {
			return err
		}

		fmt.Println("available scarlett devices:")
		for _, card := range cards {
			fmt.Printf("  %d: %s\n", card.Number, card.Name)
		}

		return nil
	},
}

var controlsCmd = &cobra.Command{
	Use:   "controls <card>",
	Short: "List all controls on a card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		controls, err := card.GetControls()
		if err != nil {
			return err
		}

		verbose, _ := cmd.Flags().GetBool("verbose")

		fmt.Printf("controls for %s:\n\n", card)
		for _, ctl := range controls {
			if verbose {
				fmt.Println(ctl.DetailedString())
			} else {
				fmt.Println(ctl.String())
			}
		}

		fmt.Printf("\ntotal: %d controls\n", len(controls))
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <card> <control-name>",
	Short: "Get the value of a control",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		ctl, err := card.FindControl(args[1])
		if err != nil {
			// Try prefix match
			ctl, err = card.FindControlByPrefix(args[1])
			if err != nil {
				return err
			}
		}

		value, err := ctl.GetValueString()
		if err != nil {
			return err
		}

		fmt.Printf("%s = %s\n", ctl.Name, value)
		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set <card> <control-name> <value>",
	Short: "Set the value of a control",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		ctl, err := card.FindControl(args[1])
		if err != nil {
			// Try prefix match
			ctl, err = card.FindControlByPrefix(args[1])
			if err != nil {
				return err
			}
		}

		err = ctl.SetValueByString(args[2])
		if err != nil {
			return err
		}

		value, _ := ctl.GetValueString()
		fmt.Printf("%s = %s\n", ctl.Name, value)
		return nil
	},
}

var routingCmd = &cobra.Command{
	Use:   "routing <card>",
	Short: "Show the current routing matrix",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		return card.PrintRoutingMatrix()
	},
}

var routeCmd = &cobra.Command{
	Use:   "route <card> <sink> <source>",
	Short: "Set a routing connection",
	Long: `Set a routing connection from a source to a sink.
Both sink and source can be specified by name or pattern.
Source can also be specified as a numeric ID.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		sinkName := args[1]
		sourceArg := args[2]

		// try to parse source as numeric ID first
		if sourceID, err := strconv.Atoi(sourceArg); err == nil {
			// find matching sink
			sinks, err := card.GetRoutingSinks()
			if err != nil {
				return err
			}

			for _, sink := range sinks {
				if strings.Contains(strings.ToLower(sink.Name), strings.ToLower(sinkName)) {
					err = sink.Control.SetValue(int64(sourceID))
					if err != nil {
						return err
					}

					value, _ := sink.Control.GetValueString()
					fmt.Printf("%s -> %s\n", sink.Name, value)
					return nil
				}
			}

			return fmt.Errorf("sink matching '%s' not found", sinkName)
		}

		// otherwise treat as source name
		err = card.SetRoutingByNames(sinkName, sourceArg)
		if err != nil {
			return err
		}

		fmt.Printf("routing updated: %s -> %s\n", sinkName, sourceArg)
		return nil
	},
}

var mixerCmd = &cobra.Command{
	Use:   "mixer <card>",
	Short: "Show the current mixer state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		return card.PrintMixerState()
	},
}

var preampCmd = &cobra.Command{
	Use:   "preamp <card>",
	Short: "Show the current preamp state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		return card.PrintPreampState()
	},
}

var watchCmd = &cobra.Command{
	Use:   "watch <card>",
	Short: "Monitor control changes in real-time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		fmt.Printf("monitoring controls for %s\n", card)

		// set up signal handler for ctrl+c
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		errChan := make(chan error, 1)

		go func() {
			errChan <- card.WatchWithDisplay()
		}()

		select {
		case <-sigChan:
			fmt.Println("\nstopping monitor...")
			return nil
		case err := <-errChan:
			return err
		}
	},
}

var gainCmd = &cobra.Command{
	Use:   "gain <card> <channel> <value>",
	Short: "Set preamp gain for a channel",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		channel, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid channel number: %s", args[1])
		}

		value, err := strconv.ParseInt(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid gain value: %s", args[2])
		}

		err = card.SetPreampGain(channel, value)
		if err != nil {
			return err
		}

		fmt.Printf("set preamp gain for channel %d to %d\n", channel, value)
		return nil
	},
}

var phantomCmd = &cobra.Command{
	Use:   "phantom <card> <channel> <on|off>",
	Short: "Set phantom power for a channel",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		card, err := scarlettctl.FindCard(args[0])
		if err != nil {
			return err
		}
		defer card.Close()

		channel, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid channel number: %s", args[1])
		}

		enabled := false
		switch strings.ToLower(args[2]) {
		case "on", "true", "1", "yes":
			enabled = true
		case "off", "false", "0", "no":
			enabled = false
		default:
			return fmt.Errorf("invalid value: %s (use on/off)", args[2])
		}

		err = card.SetPreampPhantom(channel, enabled)
		if err != nil {
			return err
		}

		state := "off"
		if enabled {
			state = "on"
		}
		fmt.Printf("set phantom power for channel %d to '%s'\n", channel, state)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(controlsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(routingCmd)
	rootCmd.AddCommand(routeCmd)
	rootCmd.AddCommand(mixerCmd)
	rootCmd.AddCommand(preampCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(gainCmd)
	rootCmd.AddCommand(phantomCmd)

	controlsCmd.Flags().BoolP("verbose", "v", false, "Show control values")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
