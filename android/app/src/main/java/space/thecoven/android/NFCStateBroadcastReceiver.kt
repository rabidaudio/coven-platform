package space.thecoven.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.internal.ChannelFlow
import kotlinx.coroutines.runBlocking

class NFCStateBroadcastReceiver: BroadcastReceiver() {

    enum class NFCState {
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
                setType(state.name)
            }
            context.sendBroadcast(i)
        }
    }

//    var handler: ((NFCState) -> Unit)? = null
    val channel = Channel<NFCState>(Channel.UNLIMITED)

    override fun onReceive(context: Context, intent: Intent) {
        val state = try {
            val type = intent.type ?: return
            NFCState.valueOf(type)
        } catch (e: IllegalArgumentException) {
            return // invalid enum
        }
//        handler?.invoke(state)
        runBlocking { channel.send(state) }
    }
}