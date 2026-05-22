package space.thecoven.android

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import space.thecoven.android.NFCStateBroadcastReceiver.NFCState

class MainActivity : FlutterActivity() {

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        val launchState = intent.getStringExtra("STATE")?.let { NFCState.valueOf(it) } ?: NFCState.unknown
        // create dependencies
        with(context.applicationContext as CovenApp) {
            nfcInterface = NFCStateInterface(flutterEngine, launchState)
        }
    }

    override fun detachFromFlutterEngine() {
        with(context.applicationContext as CovenApp) {
            nfcInterface = null
        }
        super.detachFromFlutterEngine()
    }
}