# Personas

BuzzPi serves four primary user personas. Each represents a real user type with distinct goals, technical comfort levels, and usage patterns. All design decisions are validated against these personas.

---

## Sarah — The Maker

> "I have eight Raspberry Pis doing different things and I'm tired of SSHing into each one."

**Age:** 32  
**Occupation:** Software engineer / hobbyist  
**Technical level:** High — comfortable with Linux, SSH, Docker, scripting  
**Devices:** 5-20 Raspberry Pis (Pi 4, Pi 5, Zero 2W)  

### Goals
- See all devices in one place with live status
- Execute commands (restart, update, script) across multiple devices simultaneously
- Access terminal and screen from her phone when away from her desk
- Get notified when a device goes offline or a service fails

### Pain Points
- Managing IP addresses and SSH keys across devices is tedious
- Forgetting which device does what
- No easy way to check device health from her phone
- Screen sharing requires VNC setup on every device

### Usage Pattern
- Daily: checks status, runs updates, reviews logs
- Weekly: restarts services, clears storage
- Occasional: remote access from phone when traveling
- Automations: scheduled backups, health checks

### BuzzPi for Sarah
- Workspace with terminal, screen stream, and service list
- Multi-device actions
- Notification thresholds for temperature, storage, uptime
- CLI integration for scripting

---

## Mike — The Parent

> "I want my kids to learn coding on a Raspberry Pi, but I don't want to be the IT support desk."

**Age:** 41  
**Occupation:** IT manager  
**Technical level:** Medium — comfortable with setup, avoids command line  
**Devices:** 2-5 Raspberry Pis (for kids, home automation)  

### Goals
- Set up devices for his kids and forget about them
- Remote access to help kids when they get stuck
- Ensure devices stay updated and secure without manual intervention
- Monitor storage and temperature to prevent hardware damage

### Pain Points
- Kids can't troubleshoot when something breaks
- Devices accumulate in a drawer when they stop working
- Notifications from monitoring tools are too technical ("disk usage 94%")

### Usage Pattern
- Weekly: checks that all devices are online
- Monthly: runs updates
- Ad-hoc: remote access to help kids
- Automations: automatic updates, health checks, storage alerts

### BuzzPi for Mike
- Simple device list with clear status indicators
- One-tap update all devices
- Parent-friendly notifications ("Kitchen Pi needs attention — storage is nearly full")
- Remote screen access for helping kids

---

## Ms. Chen — The Teacher

> "I have 30 Raspberry Pis in my classroom and I need to manage them between classes."

**Age:** 38  
**Occupation:** Computer science teacher  
**Technical level:** Medium — teaches Python, comfortable with Raspberry Pi setup  
**Devices:** 30 Raspberry Pis (classroom lab)  

### Goals
- Reset all devices to a known state between classes
- Push assignments and updates to all devices simultaneously
- See which devices are online and functional at a glance
- Lock devices during tests

### Pain Points
- Spending 10 minutes of every class getting devices ready
- Students breaking configurations between classes
- No way to remotely reset or reimage devices
- Managing 30 SD cards

### Usage Pattern
- Daily: check all devices are online before class
- Per class: reset device state, push assignments
- Weekly: update all devices, review storage
- Automations: classroom mode (lock settings), scheduled reset

### BuzzPi for Ms. Chen
- Device grouping (per class, per student)
- Group actions: reset, update, lock
- Classroom mode with restricted student access
- Assignment push via Extensions
- Status dashboard with class-wide health

---

## Diego — The Maker (Advanced)

> "I built a weather station with my Pi and I want to access it from anywhere without opening ports."

**Age:** 27  
**Occupation:** Electrical engineering student  
**Technical level:** High — comfortable with embedded Linux, networking, custom hardware  
**Devices:** 3-10 Raspberry Pis (personal projects, some with custom hardware)  

### Goals
- Access devices from outside his home network without VPN
- Stream sensor data from custom hardware
- Get push notifications when hardware events trigger
- Build custom Extensions for his projects

### Pain Points
- Home ISP uses CGNAT — no port forwarding possible
- VPN adds latency and complexity
- Commercial IoT platforms are expensive and lock him in

### Usage Pattern
- Daily: check sensor readings, review logs
- Weekly: tweak configurations, update firmware
- Ad-hoc: reboot after power outage
- Automations: sensor thresholds, data logging

### BuzzPi for Diego
- Relay-based connection (no port forwarding, no VPN)
- Extension SDK for custom hardware integrations
- Screen stream for headless devices
- Data streaming via BPP
- Sensor dashboard in Workspace
