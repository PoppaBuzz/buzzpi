# User Journeys

These journeys describe the complete end-to-end experience for each persona. They begin before the user has BuzzPi installed and end with the device integrated into their daily workflow.

---

## Journey 1: Sarah Sets Up Her Lab (First-Time)

> **Persona:** Sarah, the Maker  
> **Scenario:** Sarah just bought a new Pi 5 and wants to add it to BuzzPi alongside her existing 7 Pis. This is her first time using BuzzPi.

### Step 1: Discovery

Sarah hears about BuzzPi from a blog post. She visits buzzpi.dev and sees:
- "Manage all your Raspberry Pis from one place"
- "No IP addresses needed"
- "Terminal, screen, and services — on your phone"

**Feeling:** Curious, skeptical. She's tried other tools that promised simplicity.

### Step 2: Installation

Sarah downloads BuzzPi for Android. The onboarding screens show:
1. "BuzzPi finds your Pis automatically"
2. "Tap to pair — no configuration required"
3. "Access them from anywhere"

She opens the app and sees: **"No devices yet."**

**Feeling:** Neutral. She expected this.

### Step 3: Installing the Runtime

The app prompts her to install BuzzPi Runtime on her Pi. It gives her a one-line command:

```
curl -sSL https://buzzpi.dev/install | bash
```

She SSHes into her new Pi 5, pastes the command. It installs in 15 seconds. The Runtime starts automatically and prints a pairing code.

**Feeling:** Impressed. One command, no configuration files.

### Step 4: Pairing

Back in the BuzzPi app, she taps "Pair Device." The app starts scanning. Within 5 seconds, the Pi 5 appears:

```
New device found: buzzpi-5c2a
```

She taps it, enters the pairing code shown on the Pi. The device moves to "Paired" instantly. She renames it to "Workshop Pi."

**Feeling:** Delighted. This was seamless.

### Step 5: First Workspace

Sarah taps the device card. The Workspace opens with:
- **Status:** Online, IP 192.168.1.42, temp 42°C, uptime 12s
- **Services:** BuzzPi Runtime (running)
- **Storage:** 27GB free of 32GB
- **Actions:** Terminal, Screen, Restart, SSH Key

She taps Terminal. A shell prompt appears. She runs `htop` — it works.

She taps Screen. After a brief "Connecting…" the Pi's desktop appears. She watches it update as she moves the mouse on her phone.

**Feeling:** Very impressed. This is exactly what she wanted.

### Step 6: Adding More Devices

She goes to the device list. Her new Pi 5 is there. She taps the "+" icon again, re-runs the install command on each of her existing Pis (via SSH), and pairs them one by one. Each takes under a minute.

After 15 minutes, all 8 devices are in her list, each with a friendly name and live status.

**Feeling:** Accomplished. She has a dashboard for her entire lab.

### Step 7: Daily Use

A week later, Sarah opens BuzzPi on her phone while commuting. She sees "Workshop Pi" is offline. She taps it — it shows "Last seen 2 hours ago." She makes a mental note to check the power supply when she gets home.

**Feeling:** Informed, not worried. BuzzPi caught it before it became a problem.

---

## Journey 2: Mike Sets Up for His Kids

> **Persona:** Mike, the Parent  
> **Scenario:** Mike bought two Pi 5s for his kids (ages 10 and 13). He wants them to learn coding without him being the IT support.

### Step 1: Setup

Mike installs BuzzPi Runtime on both Pis and pairs them. He names them "Liam's Pi" and "Aria's Pi."

In settings, he enables:
- Automatic updates
- Storage alerts at 80%
- Temperature alerts at 70°C
- Parental controls (restrict certain actions)

He creates accounts for his kids with limited permissions: they can access their own device's workspace but cannot restart, update, or change settings.

**Feeling:** Confident. He's set it and forget it.

### Step 2: Kid Experience

Liam opens BuzzPi on his tablet. He sees only "Liam's Pi" in his device list. He taps it and gets:
- Terminal (with a welcome message)
- File browser
- Simple action buttons (Shutdown, Reboot — with confirmation)
- Status indicators

He opens the terminal and starts typing Python. It works.

**Feeling:** Excited. He's coding on his Pi from his tablet.

### Step 3: Parent Notification

A month later, Mike gets a notification: "Liam's Pi storage is 82% full." He opens BuzzPi, checks storage, sees the SD card is filling up with log files. He connects via screen, runs a log cleaner, and frees 4GB. All from his phone.

He messages Liam: "Your Pi was getting full — cleaned it up. Keep coding!"

**Feeling:** Useful. He handled it in 30 seconds without interrupting his day.

### Step 4: Troubleshooting

Aria's Pi won't boot. Mike sees it's been offline for 3 hours. He checks the power supply, reseats the SD card, and it comes back online. BuzzPi notifies him: "Aria's Pi is online."

He doesn't need to call anyone, doesn't need to Google anything. BuzzPi told him what was wrong and when it was fixed.

**Feeling:** Relieved. This would have been a frustrating evening before.

---

## Journey 3: Ms. Chen Manages Her Classroom

> **Persona:** Ms. Chen, the Teacher  
> **Scenario:** It's the start of a new semester. Ms. Chen needs to set up 30 Pis for her computer science class.

### Step 1: Bulk Setup

Ms. Chen flashes SD cards with Raspberry Pi OS + BuzzPi Runtime pre-installed. She numbers each Pi 1-30 and labels the boxes.

She opens BuzzPi on her tablet. All 30 Pis appear as they boot up. She taps each one, enters the pairing code, and assigns it a name: "Station 01" through "Station 30."

She creates a group: "CS101 — Fall." All 30 devices are added to the group.

**Feeling:** Efficient. Setup took 40 minutes for 30 devices, not 2 days.

### Step 2: Classroom Mode

Ms. Chen enables **Classroom Mode** on the group:
- Students can only access their assigned device
- Actions restricted to: open terminal, run Python, browse files
- No restart, no shutdown, no settings changes
- Screen recording is disabled

She previews what a student sees by switching to a student view.

**Feeling:** In control. She can let students explore without fear.

### Step 3: During Class

Students open BuzzPi on their phones. Each sees their assigned Pi. They open the terminal and start the lesson.

Ms. Chen opens the group dashboard:
- 30/30 devices online
- 4 devices with temperature > 60°C (she mentally notes to check placement)
- 1 device with low storage (probably a student saving large files)
- Overview of active terminal sessions

She can broadcast a message to all devices: "Class, open today's assignment with: `python3 lesson4.py`"

**Feeling:** Empowered. She has full visibility and control.

### Step 4: End of Class

Ms. Chen taps "Reset All" on the group. All 30 devices:
1. Save student work to a shared drive
2. Reset to clean state
3. Reboot

By the time the bell rings, all devices are ready for the next class.

**Feeling:** Satisfied. This saved her 10 minutes of running around the room.

---

## Journey 4: Diego Goes Remote

> **Persona:** Diego, the Advanced Maker  
> **Scenario:** Diego built a weather station with a Pi Zero 2W at a remote cabin. No static IP, CGNAT, no port forwarding.

### Step 1: The Problem

Diego has been using a mix of:
- A cron job that emails him the temperature every hour
- An old laptop that acts as a reverse SSH tunnel
- A cloud server that costs $5/month just for a VPN

None of this works reliably. The email cron job fails when the Pi's WiFi disconnects. The SSH tunnel dies after 3 days. He's looking for a simpler solution.

**Feeling:** Frustrated. He spends more time maintaining his management tools than his project.

### Step 2: BuzzPi Relay

Diego installs BuzzPi Runtime on his weather station Pi. The Runtime connects to BuzzPi Cloud via an outbound WebSocket — no inbound ports needed. It works through CGNAT, through NAT, through hotel WiFi.

He pairs the device from home. It appears as "Weather Station" in his device list, marked as "Online (relay)."

**Feeling:** Amazed. It just worked through the hardest network.

### Step 3: Data Streaming

Diego writes a small Extension for his weather station that streams sensor data via BPP. Temperature, humidity, pressure — every 30 seconds.

Back in BuzzPi, he opens the sensor dashboard for the Weather Station. Live data displays in a chart. He sets a notification: "Alert me if temperature drops below 0°C."

**Feeling:** Powerful. He built his own IoT dashboard in an afternoon.

### Step 4: Remote Access

A snowstorm knocks out power at the cabin. Diego gets a notification: "Weather Station is offline." 6 hours later: "Weather Station is online."

He taps the device, opens terminal. Checks logs — power was out, Pi rebooted when it came back, Runtime started automatically, WebSocket reconnected. Everything is fine.

He opens screen stream, watches the boot log replay. Confirms the weather station software started correctly.

**Feeling:** Confident. He trusts BuzzPi to keep his remote devices accessible.
