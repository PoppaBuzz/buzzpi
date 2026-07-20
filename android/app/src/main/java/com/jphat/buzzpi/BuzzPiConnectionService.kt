package com.jphat.buzzpi

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.data.bpp.ConnectionState
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import javax.inject.Inject

@AndroidEntryPoint
class BuzzPiConnectionService : Service() {

    @Inject
    lateinit var bppClient: BppClient

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
    private val tag = "BuzzPiConnSvc"

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification("BuzzPi Connected"))
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_CONNECT -> {
                val deviceId = intent.getStringExtra(EXTRA_DEVICE_ID) ?: return START_STICKY
                scope.launch {
                    bppClient.connectionState.collect { state ->
                        when (state) {
                            ConnectionState.CONNECTED -> {
                                updateNotification("Connected to $deviceId")
                            }
                            ConnectionState.DISCONNECTED, ConnectionState.ERROR -> {
                                updateNotification("Disconnected")
                                stopSelf()
                            }
                            else -> { }
                        }
                    }
                }
            }
            ACTION_DISCONNECT -> {
                scope.launch {
                    bppClient.disconnect()
                }
                stopSelf()
            }
        }
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private fun createNotificationChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID,
            "BuzzPi Connection",
            NotificationManager.IMPORTANCE_LOW
        ).apply {
            description = "Shows when BuzzPi is connected to a device"
            setShowBadge(false)
        }

        val manager = getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)
    }

    private fun createNotification(text: String): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("BuzzPi")
            .setContentText(text)
            .setSmallIcon(android.R.drawable.ic_menu_share)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(text: String) {
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, createNotification(text))
    }

    companion object {
        const val CHANNEL_ID = "buzzpi_connection"
        const val NOTIFICATION_ID = 1001
        const val ACTION_CONNECT = "com.jphat.buzzpi.CONNECT"
        const val ACTION_DISCONNECT = "com.jphat.buzzpi.DISCONNECT"
        const val EXTRA_DEVICE_ID = "device_id"
    }
}
