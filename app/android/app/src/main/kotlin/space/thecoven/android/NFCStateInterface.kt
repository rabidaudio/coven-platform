package space.thecoven.android

import android.util.Log
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import space.thecoven.android.NFCStateBroadcastReceiver.NFCState

class NFCStateInterface(flutterEngine: FlutterEngine) : MethodChannel.MethodCallHandler {

    private var state: NFCState = NFCState.Unknown

    private val channel: MethodChannel = MethodChannel(
        flutterEngine.dartExecutor.binaryMessenger, "space.thecoven/nfc"
    ).also {
        it.setMethodCallHandler(this)
    }

    init {
        Log.d("NFC", "NFCStateInterface created")
    }

    override fun onMethodCall(
        call: MethodCall,
        result: MethodChannel.Result
    ) {
        when (call.method) {
            "getNFCState" -> result.success(state.serialize())
            else -> result.notImplemented()
        }
    }

    fun setState(state: NFCState) {
        this.state = state
        channel.invokeMethod("onNFCState", state.serialize())
    }

    private fun NFCState.serialize() = name.lowercase()
}
