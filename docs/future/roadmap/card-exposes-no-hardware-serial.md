---
title: Card exposes no hardware serial
state: researching
created: 2026-10-07
tags: [enhancement]
source: patchbay stage 7
---

`Card` exposes a number and a name but not the interface's serial. a consumer that needs to match a card to another system's view of the same hardware (patchbay matches against PipeWire's `device.serial`) has to go around through the ALSA card number instead, which is stable only within a boot. expose the USB serial on `Card` (the kernel exposes it under the card's sysfs `usbid`/`serial`, and `alsa_card.usb-…` names carry it), so a card can be found by serial directly.
