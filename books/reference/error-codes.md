# Error Codes Reference

Complete registry of all BPP error codes with messages and recovery hints.

## 1xxx — Transport Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 1000 | CONNECTION_TIMEOUT | Connection attempt timed out | Check network and try again |
| 1001 | CONNECTION_REFUSED | Device rejected the connection | Device may be busy; try again later |
| 1002 | CONNECTION_LOST | Established connection was lost | Auto-reconnecting; check network stability |
| 1003 | CONNECTION_RATE_LIMITED | Too many connection attempts | Wait and try again |
| 1004 | TRANSPORT_UNAVAILABLE | No transport available | Device is unreachable; check device power and network |
| 1005 | RELAY_DISCONNECTED | Relay WebSocket disconnected | Auto-reconnecting |

## 2xxx — Identity Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 2000 | AUTHENTICATION_FAILED | Challenge-response verification failed | Device may have been reset; try re-pairing |
| 2001 | AUTHENTICATION_EXPIRED | Session has expired | Re-authenticate |
| 2002 | TOKEN_INVALID | Session token is invalid or revoked | Log in again |
| 2003 | TOKEN_EXPIRED | Session token has expired | Log in again |
| 2004 | PAIRING_CODE_INVALID | Pairing code is incorrect | Check the code on your device and try again |
| 2005 | PAIRING_CODE_EXPIRED | Pairing code has expired | Generate a new code on the device |
| 2006 | PAIRING_ALREADY_PAIRED | Device is already paired | Unpair from the other account first |
| 2007 | DEVICE_NOT_FOUND | Device not registered with this account | Check device ID |
| 2008 | KEY_MISMATCH | Identity key does not match stored key | Device may have been reset; re-pair |

## 3xxx — Protocol Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 3000 | INVALID_MESSAGE | Message envelope is malformed | Update your app |
| 3001 | INVALID_JSON | Message body is not valid JSON | Update your app |
| 3002 | VERSION_MISMATCH | Protocol versions are incompatible | Update your app or device Runtime |
| 3003 | METHOD_NOT_FOUND | Requested method is not available | Update your app or check device capabilities |
| 3004 | METHOD_NOT_AVAILABLE | Method exists but is not currently available | Device may be offline or feature disabled |
| 3005 | INVALID_PARAMS | Method parameters failed validation | Update your app |
| 3006 | MESSAGE_TOO_LARGE | Message exceeds size limit | Reduce request size |
| 3007 | UNSUPPORTED_COMPRESSION | Compression method not supported | Update your app |

## 4xxx — Service Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 4000 | SERVICE_ERROR | Generic service error | Check device status and try again |
| 4001 | SERVICE_NOT_AVAILABLE | Service not available on this device | Device does not support this feature |
| 4002 | SERVICE_NOT_INSTALLED | Extension not installed | Install the required extension |
| 4003 | SERVICE_BUSY | Service is busy | Try again later |
| 4004 | SERVICE_TIMEOUT | Service did not respond | Device may be overloaded; try again later |
| 4005 | SESSION_LIMIT_EXCEEDED | Too many concurrent sessions | Close other sessions and try again |
| 4006 | RATE_LIMITED | Too many requests | Wait and try again |

## 5xxx — Resource Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 5000 | NOT_FOUND | Requested resource not found | Check that the file, container, or service exists |
| 5001 | ALREADY_EXISTS | Resource already exists | Use a different name or path |
| 5002 | PERMISSION_DENIED | Client does not have required permission | Grant the permission in extension settings |
| 5003 | ACCESS_DENIED | Device denied access | Check file permissions on the device |
| 5004 | RESOURCE_EXHAUSTED | Device resource limit reached | Free up resources (close apps, delete files) |
| 5005 | STORAGE_FULL | Device storage is full | Free up space on the device |

## 6xxx — Extension Errors

| Code | Name | Message | Recovery |
|------|------|---------|----------|
| 6000 | EXTENSION_ERROR | Generic extension error | Check extension logs for details |
| 6001 | EXTENSION_NOT_INSTALLED | Extension is not installed | Install the extension from the registry |
| 6002 | EXTENSION_NOT_RUNNING | Extension is installed but not running | Start the extension |
| 6003 | EXTENSION_CRASHED | Extension process crashed | Restart the extension |
| 6004 | EXTENSION_TIMEOUT | Extension did not respond | Reload the extension |
| 6005 | EXTENSION_PERMISSION_DENIED | Extension lacks required permission | Update extension permissions |
| 6006 | PLUGIN_INVALID | Plugin manifest or binary is invalid | Reinstall the extension |
| 6007 | PLUGIN_SIGNATURE_INVALID | Plugin signature verification failed | Reinstall from a trusted source |
| 6008 | EXTENSION_UPDATE_AVAILABLE | Extension update required | Update the extension before proceeding |
