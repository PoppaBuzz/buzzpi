# Nouns

## Device

**Preferred:** Device
**Avoid:** Host, Machine, Node, Endpoint, Server, Computer, Hardware

A single Raspberry Pi (or other Linux device) running the BuzzPi Runtime. Devices are the fundamental unit of management in the platform.

**Icon:** Single-board computer outline

---

## Workspace

**Preferred:** Workspace
**Avoid:** Dashboard, Control Panel, Overview, Home, Console

The primary screen when interacting with a device. The Workspace shows the device's state, capabilities, and available actions. It is the user's starting point for every task.

**Icon:** Grid/circles layout

---

## Capability

**Preferred:** Capability
**Avoid:** Feature, Skill, Function, Ability

A discrete function a device supports. Examples: Docker, GPIO, Camera, Terminal, Screen Streaming. Capabilities are discovered at runtime and determine what the UI shows.

**Icon:** Puzzle piece or plug

---

## Action

**Preferred:** Action
**Avoid:** Command, Task, Operation, Job

A single operation performed on a device. Actions are not limited to shell commands — they include Restart Docker, Take Photo, Toggle GPIO, Update Packages.

**Icon:** Lightning bolt or play

---

## Extension

**Preferred:** Extension
**Avoid:** Plugin, Add-on, Module, Package, Plug-in

A software module that extends the BuzzPi Runtime with new capabilities. Extensions are independent processes that communicate with the Runtime over IPC. Examples: Docker Extension, Pi-hole Extension, Home Assistant Extension.

**Icon:** Puzzle piece or plus

---

## Session

**Preferred:** Session
**Avoid:** Connection, Link, Channel

An authenticated, encrypted communication channel between a client and a device. Sessions persist across network changes and can be suspended and resumed.

**Icon:** Two connected nodes

---

## Profile

**Preferred:** Profile
**Avoid:** Account, Settings, Configuration (when referring to user identity)

A user's identity and preferences within the BuzzPi Platform. Includes credentials, paired devices, Workspace layout, and notification preferences.

**Icon:** Person or avatar

---

## Organization

**Preferred:** Organization
**Avoid:** Team, Group, Workspace (when referring to multi-user)

A collection of users who share access to a set of devices. Organizations enable multi-user device management, role-based access, and shared Workspaces.

**Icon:** Building or people group

---

## Collection

**Preferred:** Collection
**Avoid:** Group, Folder, Tag, Category

A user-defined grouping of devices. Collections are personal (not shared like Organization) and can be organized by project, location, or any user-defined criteria.

**Icon:** Folder or stack

---

## Automation

**Preferred:** Automation
**Avoid:** Recipe, Rule, Script, Macro

A triggered sequence of Actions that runs automatically when conditions are met. Example: "When CPU exceeds 80%, send notification and restart Docker."

**Icon:** Circular arrows or robot

---

## Workflow

**Preferred:** Workflow
**Avoid:** Pipeline, Process, Flow

A multi-step sequence of Actions that a user can run on-demand. Unlike Automations, Workflows are manually triggered and may include user decision points.

**Icon:** Connected nodes with arrows

---

## Credential

**Preferred:** Credential
**Avoid:** Password, Key, Token, Secret

Any secret used for authentication — SSH keys, API tokens, passwords. Credentials are stored in the device's secure enclave where available.

**Icon:** Key or shield

---

## Identity

**Preferred:** Identity
**Avoid:** User, Account (when referring to cryptographic identity)

The cryptographic identity of a device or user, rooted in public/private key pairs. Identities are established during pairing and verified on every connection.

**Icon:** Fingerprint or badge

---

## Trust

**Preferred:** Trust
**Avoid:** Verification, Approval, Authorization (when referring to device relationships)

The established relationship between a client and a device after successful pairing. Trust is mutual — the client trusts the device and the device trusts the client.

**Icon:** Shield check or handshake
