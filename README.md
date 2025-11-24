# scarlettctl

a Go library and command-line tool for controlling Focusrite Scarlett, Vocaster, and Clarett USB audio interfaces on Linux.

**This is very early-stage software and has only been tested on 4th generation Scarlett interfaces!**

## overview

**scarlettctl** provides programmatic access to all mixer, routing, and preamp controls on Focusrite Scarlett devices through the Linux ALSA control interface. use it to build custom control applications, create automation scripts, or integrate Scarlett device control into your Go applications.

this project was developed by reverse engineering the [alsa-scarlett-gui](https://github.com/geoffreybennett/alsa-scarlett-gui) project by Geoffrey Bennett. see [CREDITS.md](CREDITS.md) for full attribution.

## features

- **device discovery** - automatically detect connected Scarlett/Vocaster/Clarett devices
- **control access** - read and write all ALSA controls (mixer, routing, preamp)
- **routing matrix** - view and configure complete audio routing
- **mixer control** - adjust all mixer input levels and monitor mixes
- **preamp settings** - control gain, phantom power, air mode, pad, impedance
- **real-time monitoring** - watch for control changes as they happen
- **event-driven** - subscribe to hardware and software control events

## installation

### prerequisites

install ALSA development libraries:

```bash
# debian/ubuntu
sudo apt-get install libasound2-dev

# fedora/rhel
sudo dnf install alsa-lib-devel

# arch
sudo pacman -S alsa-lib
```

ensure Go 1.21 or later is installed:

```bash
go version
```

### build from source

```bash
# clone the repository
git clone https://github.com/michaelquigley/scarlettctl.git
cd scarlettctl

# download dependencies
go mod download

# build the CLI tool
go build -o bin/scarlettctl ./cmd/scarlettctl

# optionally install system-wide
sudo install -m 755 bin/scarlettctl /usr/local/bin/
```

### permissions

add your user to the `audio` group to access ALSA devices:

```bash
sudo usermod -a -G audio $USER
```

log out and back in for the change to take effect.

## quick start

list connected devices:

```bash
scarlettctl list
```

view all controls:

```bash
scarlettctl controls 0 --verbose
```

set phantom power on channel 1:

```bash
scarlettctl phantom 0 1 on
```

view the routing matrix:

```bash
scarlettctl routing 0
```

monitor control changes in real-time:

```bash
scarlettctl watch 0
```

## CLI usage

### device commands

**list devices:**
```bash
scarlettctl list
```

example output:
```
available scarlett devices:
  0: Scarlett 18i20 USB
```

**list controls:**
```bash
# show all control names
scarlettctl controls 0

# show controls with current values
scarlettctl controls 0 --verbose
```

### control commands

**get control value:**
```bash
# by exact name
scarlettctl get 0 "Line In 1 Phantom Power Capture Switch"

# by prefix match
scarlettctl get 0 "Line In 1 Phantom"
```

**set control value:**
```bash
# boolean values: on/off, true/false, 1/0, yes/no
scarlettctl set 0 "Line In 1 Phantom Power Capture Switch" on

# integer values
scarlettctl set 0 "Line In 01 Gain Capture Volume" 128

# enumerated values (by name or index)
scarlettctl set 0 "PCM 01 Capture Enum" "Analogue 1"
scarlettctl set 0 "PCM 01 Capture Enum" 5
```

### routing commands

**view routing matrix:**
```bash
scarlettctl routing 0
```

example output:
```
routing sources
────────────────────────────────────────────────────────────
off:
  [ 0] Off                  Off

hardware inputs:
  [ 5] Analogue 1           Hardware [Analogue]
  [ 6] Analogue 2           Hardware [Analogue]
  ...

routing matrix
────────────────────────────────────────────────────────────
PCM capture (to computer/DAW):
  PCM 01 Capture Enum          <- Analogue 1        (Hardware Analogue)
  PCM 02 Capture Enum          <- Analogue 2        (Hardware Analogue)
  ...
```

**set routing:**
```bash
# by source name
scarlettctl route 0 "PCM 01" "Analogue 1"

# by source ID
scarlettctl route 0 "PCM 01" 5

# pattern matching
scarlettctl route 0 "Mixer Input 01" "Mix A"
```

### mixer commands

**view mixer state:**
```bash
scarlettctl mixer 0
```

example output:
```
mixer state:
============
Mix A:
  input 01:   127 [0..255]
  input 02:     0 [0..255]
  ...

Mix B:
  input 01:     0 [0..255]
  input 02:   200 [0..255]
  ...
```

### preamp commands

**view preamp state:**
```bash
scarlettctl preamp 0
```

example output:
```
preamp state:
=============

channel 1:
  gain:         128 [0..255]
  phantom 48v:  On
  air:          Off
  pad:          Off
  level:        Line

channel 2:
  gain:         100 [0..255]
  phantom 48v:  Off
  air:          Off
  ...
```

**set preamp gain:**
```bash
# set channel 1 gain to 128
scarlettctl gain 0 1 128
```

**control phantom power:**
```bash
# turn on phantom power for channel 1
scarlettctl phantom 0 1 on

# turn off phantom power for channel 2
scarlettctl phantom 0 2 off
```

### monitoring

**watch control changes:**
```bash
scarlettctl watch 0
```

example output (real-time updates):
```
monitoring controls for card 0: Scarlett 18i20 USB
monitoring for events... (press ctrl+c to stop)
[14:23:45] Line In 01 Gain Capture Volume          = 150
[14:23:46] Line In 1 Phantom Power Capture Switch = On
[14:23:48] PCM 01 Capture Enum                    = Analogue 3
```

press ctrl+c to stop monitoring.

## library usage

### installation

```bash
go get github.com/michaelquigley/scarlettctl
```

### import

```go
import "github.com/michaelquigley/scarlettctl"
```

### basic example

```go
package main

import (
    "fmt"
    "github.com/michaelquigley/scarlettctl"
)

func main() {
    // open card 0
    card, err := scarlettctl.OpenCard(0)
    if err != nil {
        panic(err)
    }
    defer card.Close()

    // list all controls
    controls, err := card.GetControls()
    if err != nil {
        panic(err)
    }

    fmt.Printf("found %d controls\n", len(controls))
}
```

### control operations

```go
// find a control by name
ctl, err := card.FindControl("Line In 1 Phantom Power Capture Switch")
if err != nil {
    panic(err)
}

// read current value
value, err := ctl.GetValue()
fmt.Printf("phantom power: %d\n", value)

// set value (1 = on)
err = ctl.SetValue(1)

// or use string-based setting
err = ctl.SetValueByString("on")
```

### routing operations

```go
// get all routing sources
sources, err := card.GetRoutingSources()
for _, src := range sources {
    fmt.Printf("%d: %s (%s)\n", src.ID, src.Name, src.Category)
}

// get all routing sinks
sinks, err := card.GetRoutingSinks()
for _, sink := range sinks {
    value, _ := sink.Control.GetValue()
    fmt.Printf("%s -> %s\n", sink.Name, sources[value].Name)
}

// set routing by names
err = card.SetRoutingByNames("PCM 01", "Analogue 1")

// or set by source ID
err = card.SetRouting("PCM 01 Capture Enum", 5)
```

### mixer operations

```go
// get all mixer inputs
inputs, err := card.GetMixerInputs()
for _, input := range inputs {
    value, _ := input.Control.GetValue()
    fmt.Printf("%s input %02d: %d\n", input.MixName, input.InputNum, value)
}

// set mixer level for Mix A, input 1
err = card.SetMixerLevel("Mix A", 1, 200)

// get mixer level
level, err := card.GetMixerLevel("Mix A", 1)
```

### preamp operations

```go
// get all preamp channels
channels, err := card.GetPreampChannels()
for _, ch := range channels {
    if ch.Gain != nil {
        value, _ := ch.Gain.GetValue()
        fmt.Printf("channel %d gain: %d\n", ch.ChannelNum, value)
    }
}

// set phantom power (channel 1, on)
err = card.SetPreampPhantom(1, true)

// set gain (channel 1, value 128)
err = card.SetPreampGain(1, 128)

// set air mode (channel 1, on)
err = card.SetPreampAir(1, true)

// set pad (channel 1, on)
err = card.SetPreampPad(1, true)
```

### event monitoring

```go
// create event monitor
monitor := card.NewEventMonitor()

// watch for control changes
err = monitor.WatchControls(func(control *scarlettctl.Control, value int64) error {
    fmt.Printf("%s changed to %d\n", control.Name, value)
    return nil
})

// or use the simpler display version
err = card.WatchWithDisplay()
```

## API reference

### card operations

- `OpenCard(cardNum int) (*Card, error)` - open a card by number
- `FindCard(identifier string) (*Card, error)` - find card by number or name substring
- `ListCards() ([]*Card, error)` - list all Scarlett/Vocaster/Clarett cards
- `(*Card).Close() error` - close the card connection
- `(*Card).IsScarlett() bool` - check if card is a supported device

### control operations

- `(*Card).GetControls() ([]*Control, error)` - get all controls
- `(*Card).FindControl(name string) (*Control, error)` - find by exact name
- `(*Card).FindControlByPrefix(prefix string) (*Control, error)` - find by prefix
- `(*Card).FindControlsMatching(pattern string) ([]*Control, error)` - find by substring
- `(*Control).GetValue() (int64, error)` - read control value
- `(*Control).SetValue(value int64) error` - write control value
- `(*Control).GetValueString() (string, error)` - read value as human-readable string
- `(*Control).SetValueByString(valueStr string) error` - write value from string

### routing operations

- `(*Card).GetRoutingSources() ([]RoutingSource, error)` - list all routing sources
- `(*Card).GetRoutingSinks() ([]RoutingSink, error)` - list all routing sinks
- `(*Card).GetRouting() (map[string]int, error)` - get current routing configuration
- `(*Card).SetRouting(sinkName string, sourceID int) error` - set routing by source ID
- `(*Card).SetRoutingByNames(sinkName, sourceName string) error` - set routing by names
- `(*Card).PrintRoutingMatrix() error` - display routing matrix

### mixer operations

- `(*Card).GetMixerInputs() ([]MixerInput, error)` - list all mixer inputs
- `(*Card).GetMixerInput(mixName string, inputNum int) (*Control, error)` - get specific input
- `(*Card).GetMixerLevel(mixName string, inputNum int) (int64, error)` - get input level
- `(*Card).SetMixerLevel(mixName string, inputNum int, level int64) error` - set input level
- `(*Card).PrintMixerState() error` - display mixer state

### preamp operations

- `(*Card).GetPreampChannels() ([]PreampChannel, error)` - list all preamp channels
- `(*Card).GetPreampChannel(channelNum int) (*PreampChannel, error)` - get specific channel
- `(*Card).SetPreampGain(channelNum int, gain int64) error` - set preamp gain
- `(*Card).SetPreampPhantom(channelNum int, enabled bool) error` - set phantom power
- `(*Card).SetPreampAir(channelNum int, enabled bool) error` - set air mode
- `(*Card).SetPreampPad(channelNum int, enabled bool) error` - set pad
- `(*Card).PrintPreampState() error` - display preamp state

### event operations

- `(*Card).NewEventMonitor() *EventMonitor` - create an event monitor
- `(*EventMonitor).Watch(callback func(numid uint) error) error` - watch for events
- `(*EventMonitor).WatchControls(callback func(*Control, int64) error) error` - watch with control details
- `(*EventMonitor).Stop()` - stop the event monitor
- `(*Card).WatchWithDisplay() error` - watch and display changes

## architecture

scarlettctl uses CGO to interface with the ALSA control library (`libasound`):

```
┌─────────────┐
│ scarlettctl │  CLI tool
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│   Go Library    │  public API (card, control, routing, mixer, preamp, events)
└──────┬──────────┘
       │ CGO
       ▼
┌─────────────────┐
│  ALSA Library   │  libasound (C library)
└──────┬──────────┘
       │ ioctl()
       ▼
┌─────────────────┐
│  Kernel Driver  │  snd-usb-audio modules
└──────┬──────────┘
       │ USB
       ▼
┌─────────────────┐
│    Hardware     │  Scarlett/Vocaster/Clarett device
└─────────────────┘
```

### key concepts

**ALSA controls**: each device feature (gain, phantom power, routing, etc.) is exposed as an ALSA control with a unique name and numeric ID. controls have types (boolean, integer, enumerated) and may have value ranges or option lists.

**routing matrix**: Scarlett devices use enumerated controls to configure audio routing. each sink (destination) has a control that selects which source (input) feeds it. sources include hardware inputs, PCM playback, mixer outputs, and DSP outputs.

**event monitoring**: ALSA provides event notifications when controls change (either from software or hardware). scarlettctl uses Unix polling on ALSA file descriptors to receive these events in real-time.

## control naming patterns

ALSA control names follow specific patterns that vary by device generation:

**mixer controls:**
- gen 2/3/4: `"Mix A Input 01 Playback Volume"`
- gen 1: `"Matrix 01 Mix A Playback Volume"`

**routing controls:**
- PCM capture: `"PCM 01 Capture Enum"`
- hardware outputs: `"Analogue Output 01 Playback Enum"`

**preamp controls:**
- gain: `"Line In 01 Gain Capture Volume"`
- phantom power: `"Line In 1 Phantom Power Capture Switch"`
- air mode: `"Line In 01 Air Capture Switch"`

see [CLAUDE.md](CLAUDE.md) for complete control naming documentation.

## troubleshooting

### no devices found

verify the device is connected and recognized:

```bash
# check if device is connected
lsusb | grep Focusrite

# check if ALSA sees the device
aplay -l

# verify kernel module is loaded
lsmod | grep snd_usb
```

### permission denied

add your user to the `audio` group:

```bash
sudo usermod -a -G audio $USER
```

log out and back in for the change to take effect.

### CGO build errors

ensure ALSA development headers are installed:

```bash
# debian/ubuntu
sudo apt-get install libasound2-dev

# fedora/rhel
sudo dnf install alsa-lib-devel
```

if the build still fails, try setting CGO flags manually:

```bash
export CGO_CFLAGS="-I/usr/include/alsa"
export CGO_LDFLAGS="-lasound"
go build -o bin/scarlettctl ./cmd/scarlettctl
```

### device not responding

reset the USB connection:

```bash
# find the device
lsusb | grep Focusrite

# example output: Bus 001 Device 005: ID 1235:8214 Focusrite-Novation

# unbind and rebind (replace 1-2 with your bus-port)
echo '1-2' | sudo tee /sys/bus/usb/drivers/usb/unbind
sleep 1
echo '1-2' | sudo tee /sys/bus/usb/drivers/usb/bind
```

## credits

this project was developed by reverse engineering the [alsa-scarlett-gui](https://github.com/geoffreybennett/alsa-scarlett-gui) project by **Geoffrey Bennett**. the ALSA control interface protocol and control naming patterns were analyzed from that project.

see [CREDITS.md](CREDITS.md) for complete acknowledgments of all dependencies and contributors.

## license

MIT License - see [LICENSE](LICENSE) file for details.

this is an independent implementation that interfaces with the same ALSA control API used by alsa-scarlett-gui. no code from the original project was copied.

## contributing

contributions are welcome! please:

1. fork the repository
2. create a feature branch
3. make your changes
4. add tests if applicable
5. submit a pull request

### development

see [CLAUDE.md](CLAUDE.md) for development guidelines, project structure, and coding conventions.

## related projects

- [alsa-scarlett-gui](https://github.com/geoffreybennett/alsa-scarlett-gui) - original C/GTK GUI application (inspiration for this project)
- [ALSA Library Documentation](https://www.alsa-project.org/alsa-doc/alsa-lib/)
- [Linux Sound Subsystem](https://www.kernel.org/doc/html/latest/sound/)

## support

- **issues**: report bugs or request features via [GitHub Issues](https://github.com/michaelquigley/scarlettctl/issues)
- **discussions**: ask questions in [GitHub Discussions](https://github.com/michaelquigley/scarlettctl/discussions)
- **documentation**: see [CLAUDE.md](CLAUDE.md) for technical details

---

made with appreciation for the open source audio community.
