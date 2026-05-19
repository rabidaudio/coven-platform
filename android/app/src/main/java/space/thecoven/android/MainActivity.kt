package space.thecoven.android

import android.content.IntentFilter
import android.nfc.NfcManager
import android.os.Build
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import space.thecoven.android.ui.theme.TheCovenTheme

class MainActivity : ComponentActivity() {

    val br = NFCStateBroadcastReceiver()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val manager = getSystemService(NFC_SERVICE) as NfcManager
        val adapter = manager.defaultAdapter

        Log.d("NFC", "adapter=$adapter enabled=${adapter.isEnabled}")
        // TODO if not isEnabled show error

        // TODO: if dimensions available show on UI
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            val antenna = adapter.nfcAntennaInfo
            Log.d("NFC", "w=${antenna?.deviceWidth} h=${antenna?.deviceHeight} foldable=${antenna?.isDeviceFoldable}")
            for (antenna in antenna?.availableNfcAntennas ?: emptyList()) {
                Log.d("NFC", "antenna_x=${antenna.locationX} antenna_y=${antenna.locationY}")
            }
        }
        registerReceiver(br, IntentFilter())

        // TODO br.channel.receive

        enableEdgeToEdge()
        setContent {
            TheCovenTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Greeting(
                        name = "Android",
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }

    override fun onDestroy() {
        unregisterReceiver(br)
        super.onDestroy()
    }
}

@Composable
fun Greeting(name: String, modifier: Modifier = Modifier) {
    Text(
        text = "Hello $name!",
        modifier = modifier
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    TheCovenTheme {
        Greeting("Android")
    }
}
