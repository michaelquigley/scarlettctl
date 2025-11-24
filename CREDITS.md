# credits

this project would not exist without the work of many individuals and open source projects.

## primary acknowledgment

### alsa-scarlett-gui

the primary inspiration and technical foundation for `scarlettctl` came from reverse engineering the **alsa-scarlett-gui** project by **Geoffrey Bennett**.

- **repository**: https://github.com/geoffreybennett/alsa-scarlett-gui
- **author**: Geoffrey Bennett

Geoffrey's pioneering work in creating a GUI application for Focusrite Scarlett devices provided invaluable insights into the ALSA control interface protocol, control naming patterns, and device communication methods. the `scarlettctl` project was developed by studying and understanding the control structures and interaction patterns implemented in alsa-scarlett-gui.

we are deeply grateful to Geoffrey Bennett for his exceptional work in making Scarlett device control accessible on Linux.

## system dependencies

### ALSA library (libasound)

the **Advanced Linux Sound Architecture (ALSA)** library is the core foundation that makes all device communication possible.

- **project**: ALSA Project
- **website**: https://www.alsa-project.org/
- **documentation**: https://www.alsa-project.org/alsa-doc/alsa-lib/
- **license**: LGPL

scarlettctl interfaces with ALSA via CGO for all device operations including control enumeration, value reading/writing, and real-time event monitoring.

### linux kernel sound subsystem

the **snd-usb-audio** kernel modules provide the USB audio device drivers that expose ALSA control interfaces for Scarlett, Vocaster, and Clarett devices.

- **documentation**: https://www.kernel.org/doc/html/latest/sound/
- **license**: GPL

## license

scarlettctl is released under the MIT License. we are grateful to all dependency authors who have chosen permissive licenses that enable this project to exist.

## thanks

thank you to the entire open source community for building the tools and libraries that make projects like this possible. special thanks to:

- the ALSA Project maintainers for decades of Linux audio support
- the Go team for creating an excellent systems programming language
- all contributors to the dependencies listed above

if you've contributed to any of these projects, you've contributed to scarlettctl.
