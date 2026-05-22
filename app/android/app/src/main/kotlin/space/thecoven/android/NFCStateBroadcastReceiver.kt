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
        unknown,
        connected,
        authenticating,
        unlocked,
        failedTokenExpired,
        failedOther,
        disconnected
    }

    companion object {
        fun broadcastState(context: Context, state: NFCState) {
            val i = Intent(context, NFCStateBroadcastReceiver::class.java).apply {
                putExtra("STATE", state.name)
            }
            context.sendBroadcast(i)
        }
    }

    override fun onReceive(context: Context, intent: Intent) {
        val iface = (context.applicationContext as CovenApp).nfcInterface ?: return
        val state = try {
            val stateStr = intent.getStringExtra("STATE") ?: return
            NFCState.valueOf(stateStr)
        } catch (e: IllegalArgumentException) {
            return // invalid enum
        }
        iface.setState(state)
    }
}
