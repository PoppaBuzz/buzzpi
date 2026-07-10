# BPP Chapter 14: GPIO Service

**Layer:** Services  
**Status:** Draft  
**Version:** 1.0.0

The GPIO service provides remote access to the device's General Purpose Input/Output pins. This is specific to devices with GPIO headers (Raspberry Pi, similar SBCs).

## Overview

The GPIO service allows clients to:
- Read pin states (digital and analog via ADC)
- Write pin states (digital output)
- Use PWM (Pulse Width Modulation)
- Monitor pin state changes (edge detection and interrupts)
- Access I2C, SPI, and UART peripherals

## Methods

### gpio.pin.list

List all available GPIO pins and their current configuration.

**Request:**
```json
{
  "method": "gpio.pin.list",
  "params": {}
}
```

**Response:**
```json
{
  "method": "gpio.pin.list",
  "result": {
    "chip": "gpiochip0",
    "label": "BCM2711",
    "pins": [
      {
        "pin": 2,
        "name": "SDA1",
        "current_mode": "alt0",
        "supported_modes": ["input", "output", "alt0"],
        "value": null,
        "hardware_pin": 3,
        "is_available": true
      },
      {
        "pin": 3,
        "name": "SCL1",
        "current_mode": "alt0",
        "supported_modes": ["input", "output", "alt0"],
        "value": null,
        "hardware_pin": 5,
        "is_available": true
      },
      {
        "pin": 17,
        "name": "GPIO17",
        "current_mode": "output",
        "supported_modes": ["input", "output", "pwm"],
        "value": 1,
        "hardware_pin": 11,
        "is_available": true
      }
    ]
  }
}
```

### gpio.pin.read

Read the current value of a pin.

**Request:**
```json
{
  "method": "gpio.pin.read",
  "params": {
    "pin": 17
  }
}
```

**Response:**
```json
{
  "method": "gpio.pin.read",
  "result": {
    "pin": 17,
    "value": 1,
    "mode": "output",
    "timestamp": "2026-07-07T12:00:00Z"
  }
}
```

For analog pins (ADC):

```json
{
  "method": "gpio.pin.read",
  "params": {
    "pin": "A0",
    "mode": "analog"
  }
}
```

```json
{
  "method": "gpio.pin.read",
  "result": {
    "pin": "A0",
    "value": 512,
    "voltage": 2.5,
    "mode": "analog",
    "resolution": 10
  }
}
```

### gpio.pin.write

Set the value of a pin.

**Request:**
```json
{
  "method": "gpio.pin.write",
  "params": {
    "pin": 17,
    "value": 0,
    "mode": "output"
  }
}
```

### gpio.pin.mode

Change the mode of a pin.

**Request:**
```json
{
  "method": "gpio.pin.mode",
  "params": {
    "pin": 17,
    "mode": "input",
    "pull": "down"
  }
}
```

| Mode | Description |
|------|-------------|
| `input` | Digital input |
| `output` | Digital output |
| `pwm` | Pulse Width Modulation output |
| `alt0`-`alt5` | Alternate functions (I2C, SPI, UART) |

| Pull | Description |
|------|-------------|
| `up` | Internal pull-up resistor |
| `down` | Internal pull-down resistor |
| `none` | No pull resistor |
| `off` | Disable pull resistor |

### gpio.pin.pwm

Configure PWM output on a pin.

**Request:**
```json
{
  "method": "gpio.pin.pwm",
  "params": {
    "pin": 12,
    "frequency_hz": 1000,
    "duty_cycle_percent": 50
  }
}
```

| Parameter | Range | Description |
|-----------|-------|-------------|
| `frequency_hz` | 1-1000000 | PWM frequency |
| `duty_cycle_percent` | 0-100 | Duty cycle percentage |

### gpio.pin.watch

Subscribe to pin state changes.

**Request:**
```json
{
  "method": "gpio.pin.watch",
  "params": {
    "pin": 17,
    "edge": "both"
  }
}
```

| Edge | Trigger |
|------|---------|
| `rising` | Low to high transition |
| `falling` | High to low transition |
| `both` | Any transition |
| `none` | Disable watch |

**Events:**
```json
{
  "type": "event",
  "method": "gpio.pin.event",
  "params": {
    "pin": 17,
    "value": 1,
    "edge": "rising",
    "timestamp": "2026-07-07T12:00:00.001Z"
  }
}
```

Debounce is handled by the device (configurable, default 50μs debounce time).

### gpio.i2c.read

Read from an I2C device.

**Request:**
```json
{
  "method": "gpio.i2c.read",
  "params": {
    "bus": 1,
    "address": "0x76",
    "register": "0x00",
    "length": 2
  }
}
```

### gpio.i2c.write

Write to an I2C device.

**Request:**
```json
{
  "method": "gpio.i2c.write",
  "params": {
    "bus": 1,
    "address": "0x76",
    "register": "0x01",
    "data": "0xFF",
    "encoding": "hex"
  }
}
```

### gpio.spi.transfer

Perform an SPI transfer.

**Request:**
```json
{
  "method": "gpio.spi.transfer",
  "params": {
    "bus": 0,
    "chip_select": 0,
    "mode": 0,
    "speed_hz": 1000000,
    "bits_per_word": 8,
    "data": "deadbeef",
    "encoding": "hex"
  }
}
```

## Security

| Concern | Mitigation |
|---------|------------|
| Pin damage | Runtime validates pin modes before changing (cannot set a pin in use by an alternate function) |
| Physical damage | Runtime limits PWM frequency and duty cycle to safe defaults |
| Unauthorized access | GPIO requires explicit permission extension; not enabled by default |
| Pin conflicts | Runtime tracks pin reservations to prevent conflicts between extensions |

## GPIO Extension

GPIO access requires explicit permission because it is a safety-critical feature:

1. The GPIO extension must be installed on the device
2. The client must request `gpio:*` permission
3. The user must approve the permission during extension install

Without the extension, clients receive error code `SERVICE_NOT_AVAILABLE` for GPIO methods.

## Pin Naming

GPIO pins are identified by BCM (Broadcom) pin numbers (not physical pin numbers). Physical pin mapping is device-specific and provided in the pin list response.

## Rate Limits

| Operation | Limit |
|-----------|-------|
| Pin reads | 1000/sec |
| Pin writes | 1000/sec |
| PWM config changes | 10/sec |
| I2C transfers | 100/sec |
| SPI transfers | 100/sec |
