---
title: EventMonitor Stop flag is unsynchronized
state: researching
created: 2026-10-07
tags: [defect]
source: patchbay stage 7
---

`EventMonitor.Stop` sets `running` on one goroutine while `Watch` reads it in its loop on another, with no synchronization; the race detector flags it. make `running` an atomic or close the stop channel only and have the loop select on it. related, and worth doing in the same pass: `Watch` gives the caller no signal that polling has started, so a consumer cannot distinguish "monitor running" from "Watch called"; a started channel or callback would let a consumer gate its first read on it. also note that `Watch` always calls the callback with numid 0 (the comment in `events.go` says so); either deliver the real numid or document that consumers must refresh everything per event.
