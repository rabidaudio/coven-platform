package space.thecoven.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.internal.ChannelFlow
import kotlinx.coroutines.runBlocking

class NFCStateBroadcastReceiver: BroadcastReceiver() {

    enum class NFCState {
        Unknown,
        Connected,
        Authenticating,
        Unlocked,
        FailedTokenExpired,
        FailedOther,
        Disconnected
    }

    companion object {
        fun broadcastState(context: Context, state: NFCState) {
            val i = Intent(context, NFCStateBroadcastReceiver::class.java).apply {
                type = state.name
            }
            context.sendBroadcast(i)
        }
    }

    override fun onReceive(context: Context, intent: Intent) {
        val iface = (context.applicationContext as CovenApp).nfcInterface
        Log.d("NFC", "onReceive $intent iface=$iface")
        val state = try {
            val type = intent.type ?: return
            NFCState.valueOf(type)
        } catch (e: IllegalArgumentException) {
            return // invalid enum
        }
        iface?.setState(state)
    }
}
