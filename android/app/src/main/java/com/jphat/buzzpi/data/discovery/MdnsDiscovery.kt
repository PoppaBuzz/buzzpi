package com.jphat.buzzpi.data.discovery

import android.content.Context
import android.net.nsd.NsdManager
import android.net.nsd.NsdServiceInfo
import android.net.wifi.WifiManager
import android.util.Log
import com.jphat.buzzpi.domain.model.Device
import com.jphat.buzzpi.domain.model.Transport
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

class MdnsDiscovery(private val context: Context) {

    companion object {
        private const val TAG = "MdnsDiscovery"
        private const val SERVICE_TYPE = "_buzzpi._tcp"
        private const val DEFAULT_PORT = 8420
    }

    private var nsdManager: NsdManager? = null
    private var activeDiscovery: NsdManager.DiscoveryListener? = null
    private var multicastLock: WifiManager.MulticastLock? = null
    private val _discoveredDevices = MutableStateFlow<List<Device>>(emptyList())
    val discoveredDevices: Flow<List<Device>> = _discoveredDevices.asStateFlow()
    private val discoveredServices = mutableMapOf<String, NsdServiceInfo>()

    private fun acquireMulticastLock() {
        try {
            val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as? WifiManager
            multicastLock = wifiManager?.createMulticastLock("buzzpi_mdns")?.apply {
                setReferenceCounted(true)
                acquire()
            }
            Log.d(TAG, "Multicast lock acquired: ${multicastLock != null}")
        } catch (e: Exception) {
            Log.w(TAG, "Failed to acquire multicast lock: ${e.message}")
        }
    }

    private fun releaseMulticastLock() {
        try {
            multicastLock?.release()
            multicastLock = null
        } catch (e: Exception) {
            Log.w(TAG, "Failed to release multicast lock: ${e.message}")
        }
    }

    fun startDiscovery() {
        stopDiscovery()
        acquireMulticastLock()

        nsdManager = context.getSystemService(Context.NSD_SERVICE) as? NsdManager
        if (nsdManager == null) {
            Log.w(TAG, "NSD not available on this device")
            return
        }

        val listener = object : NsdManager.DiscoveryListener {
            override fun onDiscoveryStarted(regType: String) {
                Log.d(TAG, "mDNS discovery started: $regType")
            }

            override fun onServiceFound(serviceInfo: NsdServiceInfo) {
                Log.d(TAG, "Service found: ${serviceInfo.serviceName} (${serviceInfo.serviceType})")
                val foundType = serviceInfo.serviceType?.trimEnd('.')
                if (foundType != SERVICE_TYPE) return

                @Suppress("DEPRECATION")
                nsdManager?.resolveService(serviceInfo, object : NsdManager.ResolveListener {
                    override fun onResolveFailed(serviceInfo: NsdServiceInfo, errorCode: Int) {
                        Log.w(TAG, "Resolve failed: $errorCode for ${serviceInfo.serviceName}")
                    }

                    override fun onServiceResolved(serviceInfo: NsdServiceInfo) {
                        val device = parseServiceInfo(serviceInfo)
                        if (device != null) {
                            discoveredServices[device.deviceId] = serviceInfo
                            val current = _discoveredDevices.value.toMutableList()
                            current.removeAll { it.deviceId == device.deviceId }
                            current.add(device)
                            _discoveredDevices.value = current
                            Log.d(TAG, "Device resolved: ${device.friendlyName} @ ${device.ipAddress}")
                        }
                    }
                })
            }

            override fun onServiceLost(serviceInfo: NsdServiceInfo) {
                Log.d(TAG, "Service lost: ${serviceInfo.serviceName}")
                val deviceId = serviceInfo.serviceName
                discoveredServices.remove(deviceId)
                val current = _discoveredDevices.value.toMutableList()
                current.removeAll { it.deviceId == deviceId }
                _discoveredDevices.value = current
            }

            override fun onDiscoveryStopped(regType: String) {
                Log.d(TAG, "Discovery stopped: $regType")
            }

            override fun onStartDiscoveryFailed(regType: String, errorCode: Int) {
                Log.e(TAG, "Start discovery failed: $errorCode for $regType")
            }

            override fun onStopDiscoveryFailed(regType: String, errorCode: Int) {
                Log.e(TAG, "Stop discovery failed: $errorCode for $regType")
            }
        }

        activeDiscovery = listener
        nsdManager?.discoverServices(SERVICE_TYPE, NsdManager.PROTOCOL_DNS_SD, listener)
    }

    fun stopDiscovery() {
        activeDiscovery?.let { listener ->
            try {
                nsdManager?.stopServiceDiscovery(listener)
            } catch (e: Exception) {
                Log.w(TAG, "Error stopping discovery: ${e.message}")
            }
        }
        activeDiscovery = null
        releaseMulticastLock()
    }

    @Suppress("DEPRECATION")
    private fun parseServiceInfo(info: NsdServiceInfo): Device? {
        val host = info.host ?: return null
        val port = info.port.takeIf { it > 0 } ?: DEFAULT_PORT
        val ipAddress = host.hostAddress ?: return null
        val txtRecords = parseTxtRecords(info)

        return Device(
            deviceId = txtRecords["device_id"] ?: info.serviceName,
            friendlyName = txtRecords["friendly_name"] ?: info.serviceName,
            platform = txtRecords["platform"] ?: "",
            runtimeVersion = txtRecords["runtime_version"] ?: txtRecords["version"] ?: "",
            capabilities = (txtRecords["capabilities"] ?: "").split(",").filter { it.isNotBlank() },
            isOnline = true,
            transport = Transport.LAN,
            ipAddress = ipAddress,
            port = port
        )
    }

    private fun parseTxtRecords(info: NsdServiceInfo): Map<String, String> {
        val records = mutableMapOf<String, String>()
        try {
            val txtRecords = info.attributes
            txtRecords?.forEach { (key, value) ->
                val strValue = String(value, Charsets.UTF_8)
                records[key] = strValue
            }
        } catch (e: Exception) {
            Log.w(TAG, "Error parsing TXT records: ${e.message}")
        }
        return records
    }

    fun cleanup() {
        stopDiscovery()
        nsdManager = null
    }
}
