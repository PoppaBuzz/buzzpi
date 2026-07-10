# BuzzPi Manifesto

Computers have IP addresses. People have Raspberry Pis.

That one sentence influences every decision we make.

---

## The Nine Principles

### 1. No IP Addresses

The user should never have to know an IP address. Ever. No exceptions.

If an IP address appears on screen, we have failed. Discovery, connection, reconnection — all of it must work without the user ever seeing or entering a numeric address. This is not a preference. It is a hard requirement.

### 2. Best Connection Always

The app should always choose the best available connection automatically.

LAN, WAN, VPN, Relay, SSH, WebRTC — BuzzPi figures it out. Users should never have to think about how they are connected or switch modes manually. The Connection Engine exists so they do not have to.

### 3. Native Experience

Everything should feel native. Not a web page wrapped in chrome. Not Electron. Not Flutter. Not React Native.

Pure Android. Jetpack Compose. Material 3. Every animation, every gesture, every transition should feel like it belongs on the device. BuzzPi should feel like it was designed by the platform team, not ported from somewhere else.

### 4. Discoverability

Everything should be discoverable. No hidden menus. No terminal commands. No memorizing arcane options.

The interface grows with the user's experience. Beginners see what they need. Experts can drill deeper. But nothing is buried behind a configuration file or a long-press gesture that no one will find.

### 5. Beauty

Most Pi utilities look functional. BuzzPi should look like something Google shipped.

Beauty is not a luxury. It is a signal of quality. A beautiful interface communicates that care was taken, that the developers respected the user enough to polish every corner. BuzzPi is usable at first glance and delightful at every glance after.

### 6. Performance

60 frames per second. Instant startup. Small APK. Low battery usage. Minimal RAM footprint. Fast reconnect.

Performance is a feature, not an afterthought. Every millisecond of latency and every kilobyte of memory is accounted for. BuzzPi runs on the same hardware the user already has, without demanding upgrades.

### 7. Offline-First

If the internet dies, the app should still work on the local network.

The cloud is a fallback, not a requirement. Local discovery, local connections, local everything work without an internet connection. Remote access adds capability but never subtracts reliability.

### 8. Privacy-First

No telemetry by default. No analytics. No ads. No tracking. No data collection.

Privacy is not a feature we add later. It is the default state. Users do not need to opt out of anything because nothing is collected unless the user explicitly chooses to share it for debugging or support.

### 9. Open Everything

Android app. Go Agent. Backend. Website. Documentation. Brand assets. Plugin SDK. Everything.

BuzzPi is open source not because it is free, but because openness builds trust, attracts contributors, and ensures the project outlives any single person or company. We build in public.

---

## The BuzzPi Experience Guidelines (BEG)

Beyond principles, these guidelines shape every design decision:

**Discovery over Configuration.** Never ask the user to configure something that can be discovered automatically. If BuzzPi can detect it, BuzzPi should detect it.

**Automation over Instructions.** Never tell the user how to perform a repetitive task. Automate it. If the same action is needed more than once, it should happen without user intervention.

**Explain, Don't Expose.** Never show a raw error code. Translate everything into plain language. "SSH Error Code 255" becomes "BuzzPi could not reach your Raspberry Pi. It may be asleep, disconnected from the network, or powered off."

**Progressive Disclosure.** Beginners see "Restart." Experts can tap to expand and see "sudo systemctl restart jellyfin." The interface grows with the user's expertise. No one is overwhelmed, and no one is held back.

**Everything Has a Reason.** Every button, every screen, every notification must answer the question "Why does this exist?" If we cannot answer that, it probably should not exist.

---

## RFC-Driven Development

Every major subsystem is designed through an RFC (Request for Comments) process before code is written. RFCs are engineering documents that explain:

- Motivation — why this change matters
- Architecture — how it works
- Alternatives considered — what else was explored
- Tradeoffs — what was sacrificed and why
- Security implications — how safety is maintained
- Migration path — how existing users are affected

No major feature is implemented before its RFC is accepted. This is how we avoid redesigning major pieces later and how we ensure every contributor understands why decisions were made.

---

## What We Are Building

BuzzPi is not a remote desktop. It is not an SSH client. It is an operating system for managing Raspberry Pis, built around a protocol that any device can implement.

We are building the experience that Raspberry Pi has always deserved: effortless, beautiful, and completely free of IP addresses.
