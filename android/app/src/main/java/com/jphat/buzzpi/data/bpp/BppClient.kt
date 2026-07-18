package com.jphat.buzzpi.data.bpp

import android.util.Log
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.TimeUnit

enum class ConnectionState {
    DISCONNECTED,
    CONNECTING,
    CONNECTED,
    RECONNECTING,
    ERROR
}

class BppClient(
    private val okHttpClient: OkHttpClient = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS)
        .pingInterval(30, TimeUnit.SECONDS)
        .build()
) {
    private var webSocket: WebSocket? = null
    private var currentUrl: String = ""
    private val _connectionState = MutableStateFlow(ConnectionState.DISCONNECTED)
    val connectionState: Flow<ConnectionState> = _connectionState.asStateFlow()
    private val _messages = Channel<BppEnvelope>(Channel.BUFFERED)
    val messages: Flow<BppEnvelope> = _messages.receiveAsFlow()
    private val pendingRequests = ConcurrentHashMap<String, Channel<BppEnvelope>>()
    private var messageCounter = 0L

    companion object {
        private const val TAG = "BppClient"
        private val dateFormat: ThreadLocal<SimpleDateFormat> = ThreadLocal.withInitial {
            SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.US).apply {
                timeZone = TimeZone.getTimeZone("UTC")
            }
        }
    }

    fun isConnected(): Boolean {
        return webSocket != null && _connectionState.value == ConnectionState.CONNECTED
    }

    fun connect(url: String) {
        val state = _connectionState.value
        Log.d(TAG, "connect() url=$url currentUrl=$currentUrl state=$state")
        if (url == currentUrl && state != ConnectionState.DISCONNECTED && state != ConnectionState.ERROR) {
            Log.d(TAG, "connect() guard: skip (same url, state=$state)")
            return
        }
        Log.d(TAG, "connect() calling disconnect() then connecting")
        disconnect()
        currentUrl = url
        _connectionState.value = ConnectionState.CONNECTING

        val request = Request.Builder()
            .url(url)
            .build()

        webSocket = okHttpClient.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.d(TAG, "onOpen url=$url")
                _connectionState.value = ConnectionState.CONNECTED
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val envelope = parseEnvelope(text)
                    val rid = envelope.rid
                    Log.d(TAG, "onMessage method=${envelope.method} rid=$rid type=${envelope.type} id=${envelope.id}")
                    if (rid.isNotEmpty() && pendingRequests.containsKey(rid)) {
                        Log.d(TAG, "onMessage routing to pending request rid=$rid")
                        pendingRequests[rid]?.trySend(envelope)
                    } else {
                        Log.d(TAG, "onMessage routing to messages channel")
                        _messages.trySend(envelope)
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "onMessage parse error: ${e.message}")
                }
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.d(TAG, "onClosing code=$code reason=$reason")
                webSocket.close(1000, null)
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.d(TAG, "onClosed code=$code reason=$reason")
                _connectionState.value = ConnectionState.DISCONNECTED
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e(TAG, "onFailure: ${t.message}", t)
                _connectionState.value = ConnectionState.ERROR
            }
        })
    }

    fun disconnect() {
        Log.d(TAG, "disconnect() pendingRequests=${pendingRequests.size} url=$currentUrl", Throwable("disconnect stacktrace"))
        webSocket?.close(1000, "Client closing")
        webSocket = null
        currentUrl = ""
        _connectionState.value = ConnectionState.DISCONNECTED
        pendingRequests.forEach { (rid, channel) ->
            Log.d(TAG, "disconnect() closing channel rid=$rid")
            channel.close()
        }
        pendingRequests.clear()
    }

    /**
     * Sends a BPP request and waits for a correlated response.
     * The envelope uses the Go agent wire format: id, rid, ts fields.
     * Response correlation is done via rid (request ID match).
     */
    suspend fun sendRequest(method: String, params: JSONObject? = null): BppEnvelope {
        val msgId = nextMessageId()
        val reqId = nextRequestId()
        val timestamp = dateFormat.get()!!.format(Date())

        val envelope = mapOf(
            "v" to BppProtocol.VERSION,
            "id" to msgId,
            "ts" to timestamp,
            "type" to BppProtocol.TYPE_REQUEST,
            "method" to method,
            "rid" to reqId,
            "params" to (params ?: JSONObject())
        )

        val channel = Channel<BppEnvelope>(1)
        pendingRequests[reqId] = channel

        val jsonText = JSONObject(envelope).toString()
        val sent = webSocket?.send(jsonText) ?: false
        Log.d(TAG, "sendRequest method=$method rid=$reqId sent=$sent pendingCount=${pendingRequests.size}")
        val response = channel.receive()
        Log.d(TAG, "sendRequest received response rid=${response.rid} method=${response.method} type=${response.type} error=${response.error?.message}")
        return response
    }

    /**
     * Sends a BPP event (fire-and-forget message with no response correlation).
     * Uses the same envelope format as SendRequest but with type "event" and no rid.
     */
    fun sendMessage(method: String, params: JSONObject? = null) {
        val msgId = nextMessageId()
        val timestamp = dateFormat.get()!!.format(Date())

        val envelope = mapOf(
            "v" to BppProtocol.VERSION,
            "id" to msgId,
            "ts" to timestamp,
            "type" to BppProtocol.TYPE_EVENT,
            "method" to method,
            "params" to (params ?: JSONObject())
        )

        val jsonText = JSONObject(envelope).toString()
        webSocket?.send(jsonText)
    }

    private fun parseEnvelope(json: String): BppEnvelope {
        val obj = JSONObject(json)
        val id = obj.optString("id", "")
        val rid = obj.optString("rid", "")

        val params = if (obj.has("params") && !obj.isNull("params")) {
            obj.get("params").toString().encodeToByteArray()
        } else null

        val result = if (obj.has("result") && !obj.isNull("result")) {
            obj.get("result").toString().encodeToByteArray()
        } else null

        val error = if (obj.has("error") && !obj.isNull("error")) {
            val err = obj.getJSONObject("error")
            BppError(
                code = err.optString("code", "unknown"),
                message = err.optString("message", "")
            )
        } else null

        return BppEnvelope(
            version = obj.optInt("v", 1),
            id = id,
            ts = obj.optString("ts", ""),
            type = obj.optString("type", ""),
            method = obj.optString("method", ""),
            rid = rid,
            params = params,
            result = result,
            error = error
        )
    }

    private var msgSeq = 0L
    private var reqSeq = 0L

    private fun nextMessageId(): String {
        return "msg_${++msgSeq}"
    }

    private fun nextRequestId(): String {
        return "req_${++reqSeq}"
    }

    fun parseJsonToBytes(json: String): ByteArray = json.encodeToByteArray()
    fun bytesToJsonString(bytes: ByteArray): String = bytes.decodeToString()
}
