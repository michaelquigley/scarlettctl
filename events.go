package scarlettctl

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// EventMonitor monitors ALSA control events
type EventMonitor struct {
	card     *Card
	running  bool
	stopChan chan struct{}
}

// NewEventMonitor creates a new event monitor for the card
func (c *Card) NewEventMonitor() *EventMonitor {
	return &EventMonitor{
		card:     c,
		stopChan: make(chan struct{}),
	}
}

// Watch starts monitoring for control changes and calls the callback for each change
// The callback receives the numid of the changed control
func (em *EventMonitor) Watch(callback func(numid uint) error) error {
	if em.card.handle == nil {
		return fmt.Errorf("card not open")
	}

	em.running = true
	defer func() { em.running = false }()

	pollFds := em.card.GetPollFds()
	if len(pollFds) == 0 {
		return fmt.Errorf("no poll descriptors available")
	}

	// build pollfd array for unix.Poll
	fds := make([]unix.PollFd, len(pollFds))
	for i, fd := range pollFds {
		fds[i] = unix.PollFd{
			Fd:     int32(fd),
			Events: unix.POLLIN,
		}
	}

	fmt.Println("monitoring for events... (press ctrl+c to stop)")

	for em.running {
		// check if we should stop
		select {
		case <-em.stopChan:
			return nil
		default:
		}

		// poll with timeout
		n, err := unix.Poll(fds, 1000) // 1 second timeout
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return fmt.Errorf("poll failed: %v", err)
		}

		if n == 0 {
			// timeout, continue
			continue
		}

		// check for events
		for {
			hasEvent, err := checkEvent(em.card.handle)
			if err != nil {
				return fmt.Errorf("check event failed: %v", err)
			}

			if !hasEvent {
				break
			}

			// we detected an event but don't have the numid from checkEvent
			// for simplicity, we'll call the callback with 0
			// in a production implementation, you'd extract the numid from the event
			if callback != nil {
				if err := callback(0); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// WatchControls monitors specific controls and calls the callback with control details
func (em *EventMonitor) WatchControls(callback func(control *Control, value int64) error) error {
	// get all controls once at the start
	controls, err := em.card.GetControls()
	if err != nil {
		return err
	}

	// build a map of numid -> control for quick lookup
	controlMap := make(map[uint]*Control)
	for _, ctl := range controls {
		controlMap[ctl.NumID] = ctl
	}

	return em.Watch(func(numid uint) error {
		// if numid is 0, check all controls (since we simplified the event handling)
		// in practice, you'd check only the changed control
		for _, ctl := range controls {
			value, err := ctl.GetValue()
			if err != nil {
				continue // skip controls we can't read
			}

			if callback != nil {
				if err := callback(ctl, value); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Stop stops the event monitor
func (em *EventMonitor) Stop() {
	em.running = false
	close(em.stopChan)
}

// WatchWithDisplay monitors controls and displays changes in a human-readable format
func (c *Card) WatchWithDisplay() error {
	monitor := c.NewEventMonitor()

	lastUpdate := make(map[uint]int64)

	return monitor.WatchControls(func(control *Control, value int64) error {
		// only print if value changed
		key := control.NumID
		if lastValue, exists := lastUpdate[key]; exists && lastValue == value {
			return nil
		}

		lastUpdate[key] = value

		// format the output
		timestamp := time.Now().Format("15:04:05")
		valueStr, _ := control.GetValueString()

		fmt.Printf("[%s] %-50s = %s\n", timestamp, control.Name, valueStr)

		return nil
	})
}
