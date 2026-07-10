# Verbs

These verbs are the canonical action words for the BuzzPi Platform. Every button label, menu item, API endpoint, and documentation string uses these verbs.

---

## Pair

**Preferred:** Pair
**Avoid:** Register, Connect, Authorize, Link, Bind, Associate, Sync

Establish trust between a client and a device for the first time. Pairing is a one-time operation that creates a permanent trust relationship.

UI: "Pair Device" — "Pairing..." — "Paired Successfully"

---

## Discover

**Preferred:** Discover
**Avoid:** Scan, Search, Lookup, Resolve, Find, Probe

Locate nearby or remote devices automatically without user configuration.

UI: "Discovering Devices..." — "2 Devices Discovered"

---

## Open

**Preferred:** Open
**Avoid:** Launch, Start (when referring to Workspace), Enter, Go To

Enter a device's Workspace.

UI: "Open Workspace" — "Open Terminal" — "Open File Manager"

---

## Stream

**Preferred:** Stream
**Avoid:** View, Watch, Mirror, Cast

Start a real-time data flow — screen frames, camera feed, logs.

UI: "Stream Screen" — "Streaming..." — "Stop Stream"

---

## Monitor

**Preferred:** Monitor
**Avoid:** Watch, Track, Observe

Continuously observe a metric or state over time.

UI: "Monitor CPU" — "Monitoring" — "Stop Monitoring"

---

## Execute

**Preferred:** Execute
**Avoid:** Run, Perform, Do, Invoke

Run a single Action on a device.

UI: "Execute" — "Executing..." — "Executed"

---

## Observe

**Preferred:** Observe
**Avoid:** Inspect, Examine, View (when referring to logs or data)

View real-time data without interacting. Observing a log stream or metric differs from Monitoring because it is a single view rather than a continuous alert.

UI: "Observe Logs" — "Observing..."

---

## Automate

**Preferred:** Automate
**Avoid:** Schedule, Script, Configure Automation

Create or enable an Automation.

UI: "Automate" — "Automation Created"

---

## Extend

**Preferred:** Extend
**Avoid:** Install Plugin, Add Extension, Enable Module

Add a new Capability to the Runtime by installing an Extension.

UI: "Extend Runtime" — "Extensions" — "Install Extension"

---

## Share

**Preferred:** Share
**Avoid:** Invite, Grant Access (when referring to device access with another user)

Give another user access to a device or Collection.

UI: "Share Device" — "Shared with 3 Users"

---

## Sync

**Preferred:** Sync
**Avoid:** Transfer, Copy, Mirror, Replicate (when referring to data between devices)

Synchronize data, files, or configuration between devices.

UI: "Sync Files" — "Syncing..." — "Synced"

---

## Verb Usage Map

| UI Context | Canonical Verb |
|------------|---------------|
| First-time setup | Pair |
| Finding devices | Discover |
| Entering workspace | Open |
| Screen sharing | Stream |
| Watching metrics | Monitor |
| Running a task | Execute |
| Viewing logs | Observe |
| Setting up automation | Automate |
| Installing extension | Extend |
| Granting access | Share |
| Copying data between devices | Sync |
