package space.thecoven.android

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine

class MainActivity : FlutterActivity() {

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        (context.applicationContext as CovenApp).nfcInterface = NFCStateInterface(flutterEngine)
    }

    override fun detachFromFlutterEngine() {
        (context.applicationContext as CovenApp).nfcInterface = null
        super.detachFromFlutterEngine()
    }
}