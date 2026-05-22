package space.thecoven.android

import android.content.Intent
import android.content.Intent.FLAG_ACTIVITY_NEW_TASK
import android.nfc.cardemulation.HostApduService
import android.os.Bundle
import android.util.Base64
import android.util.Log
import android.widget.Toast
import space.thecoven.android.NFCStateBroadcastReceiver.NFCState

/**
 * This class uses Android's support for Host Card Emulation to exchange keys with the door lock
 * even from the background. Communication is handled through standards ISO-14443-4 (data-link)
 * and ISO-7816-4 (application).
 *
 * NOTE: the implementation goes well above and beyond what is strictly necessary to implement
 * the protocol. I've done so intentionally do document the ISO-7816 protocol.
 *
 * See: https://developer.android.com/develop/connectivity/nfc/hce
 * See: https://www.freecalypso.org/pub/GSM/ISO7816/ISO_7816-4_2005.pdf
 */
class DoorAccessService : HostApduService() {

    companion object {
        const val DOOR_UNLOCK_RESULT_CMD = 0xFA
    }

    override fun onCreate() {
        super.onCreate()
        Log.d("NFC", "DoorAccessService started")
    }

    override fun processCommandApdu(commandApdu: ByteArray, extras: Bundle?): ByteArray {
        Log.d("NFC", "processCommandApdu ${commandApdu.toHexString()} extras=$extras")
        // this command blocks the main thread. If it can't be executed immediately,
        // return null and call sendResponseApdu() when ready
        val res = try {
            val apdu = IDCard.CommandAPDU.decode(commandApdu)
            when (apdu) {
                is IDCard.CommandAPDU.Proprietary ->
                    Log.d("NFC", "RequestADPU ${apdu.cla.toUByte()} (proprietary) ${apdu.raw.toHexString()}")

                is IDCard.CommandAPDU.InterIndustry ->
                    Log.d(
                        "NFC",
                        "RequestADPU ${apdu.command.name} len=${apdu.data.size}"
                    )
            }
            handleCommand(apdu)
        } catch (e: IDCard.InvalidMessageException) {
            Log.e("NFC", "Invalid message", e)
            IDCard.Status.InvalidLength.toResponse()
        }
        Log.d("NFC", "ResponseADPU status=${res.status} || ${res.encode().toHexString()}")
        return res.encode()
    }

    fun handleCommand(cmd: IDCard.CommandAPDU): IDCard.ResponseAPDU {
        when (cmd) {
            is IDCard.CommandAPDU.InterIndustry -> {
                if (cmd.channel != IDCard.LogicalChannel.ZERO)
                    return IDCard.Status.UnsupportedCLAFunctionLogicalChannel.toResponse()
                if (cmd.chaining.toInt() != 0)
                    return IDCard.Status.UnsupportedCLAFunctionCommandChanning.toResponse()
                if (cmd.secureMessaging.toInt() != 0)
                    return IDCard.Status.UnsupportedCLAFunctionSecureMessaging.toResponse()

                when (cmd.command) {
                    IDCard.Command.SELECT -> {

                        val aid = getString(R.string.aid).hexToByteArray()
                        if (!cmd.data.contentEquals(aid))
                            return IDCard.Status.FileOrApplicationNotFound.toResponse()
                        setState(NFCState.connected)
                        return IDCard.Status.OK.toResponse()
                    }

                    IDCard.Command.GENERAL_AUTHENTICATE -> {
                        setState(NFCState.authenticating)
                        val challenge = cmd.data
                        Log.d("NFC", "challenge=${challenge.toHexString()}")

                        val doorPubSigningKey = getString(R.string.door_signing_pubkey).hexToByteArray()
                        val info = getString(R.string.info).toByteArray()
                        val auth = Authenticator(doorPubSigningKey, info)

                        val userTokenString = TokenManager(applicationContext).getToken()
                        if (userTokenString == null) {
                            setState(NFCState.failedTokenExpired)
                            openApp(NFCState.failedTokenExpired)
                            return IDCard.Status(0x6A.toUByte(), 0x88.toUByte()).toResponse()
                        }
                        val userToken = Base64.decode(userTokenString, Base64.DEFAULT)
                        try {
                            val encryptedToken = auth.verifyChallengeAndEncryptToken(challenge, userToken)
                            return IDCard.ResponseAPDU(data = encryptedToken)
                        } catch (e: Exception) {
                            Log.e("NFC", "Authentication Failed", e)
                            // Return a security error
                            return IDCard.Status(0x66.toUByte(), 0x00.toUByte()).toResponse()
                        }
                    }

                    else ->
                        return IDCard.Status.UnsupportedCommand.toResponse()
                }
            }

            is IDCard.CommandAPDU.Proprietary -> {
                when (val proprietaryCommand = cmd.cla.toUByte().toInt()) {
                    DOOR_UNLOCK_RESULT_CMD -> {
                        // auth result
                        if (cmd.raw.size < 7) {
                            Log.d("NFC", "invalid command size=${cmd.raw.size}")
                            return IDCard.Status.InvalidLength.toResponse()
                        }
                        val status = IDCard.Status(cmd.raw[4].toUByte(), cmd.raw[5].toUByte())
                        if (status.isOkay()) {
                            Log.d("NFC", "door unlocked")
                            setState(NFCState.unlocked)
                        } else if (status.asShort() == 0x6645.toShort()) { // expired
                            Log.d("NFC", "door NOT unlocked: expired ($status)")
                            TokenManager(applicationContext).clearToken()
                            setState(NFCState.failedTokenExpired)
                            openApp(NFCState.failedTokenExpired)
                        } else {
                            setState(NFCState.failedOther)
                            Log.d("NFC", "door NOT unlocked: $status")
                        }
                        return IDCard.Status.OK.toResponse()
                    }
                    0xFF -> {
                        // Dummy test command
                        return IDCard.ResponseAPDU(data = listOf(0xDE, 0xAD, 0xBE, 0xEF).map { it.toByte() }.toByteArray())
                    }

                    else -> {
                        Log.d("NFC", "unknown command $proprietaryCommand")
                        return IDCard.Status.UnsupportedCommand.toResponse()
                    }
                }
            }
        }
    }
    fun openApp(state: NFCState) {
        startActivity(Intent(this, MainActivity::class.java).apply {
            flags = FLAG_ACTIVITY_NEW_TASK
            putExtra("STATE", state.name)
        })
    }

    override fun onDeactivated(reason: Int) {
        setState(NFCState.disconnected)
        when (reason) {
            DEACTIVATION_LINK_LOSS ->
                Log.w("NFC", "DoorAccessService deactivated reason=link lost")

            DEACTIVATION_DESELECTED ->
                Log.w("NFC", "DoorAccessService deactivated reason=deactivated")

            else ->
                Log.w("NFC", "DoorAccessService deactivated reason=unknown ($reason)")
        }
    }

    fun setState(state: NFCState) {
        NFCStateBroadcastReceiver.broadcastState(this, state)
    }
}
