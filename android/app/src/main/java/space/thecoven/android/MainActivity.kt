package space.thecoven.android

import android.content.Context
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
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.lifecycle.DefaultLifecycleObserver
import androidx.lifecycle.LifecycleObserver
import androidx.lifecycle.LifecycleOwner
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.lifecycleScope
import androidx.lifecycle.repeatOnLifecycle
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.consumeAsFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import space.thecoven.android.ui.theme.TheCovenTheme

class MainActivity : ComponentActivity() {

    class StateViewModel : ViewModel() {
//        var state by mutableStateOf("unknown")

        val br = NFCStateBroadcastReceiver()

        val state = br.channel.consumeAsFlow()
            .map { it.name }
            .stateIn(
            viewModelScope,
            started = SharingStarted.WhileSubscribed(5_000),
            initialValue = "unknown"
        )
        fun register(context: ComponentActivity) {
//            br.handler = { state = it.name }
            context.registerReceiver(br, IntentFilter())
            context.lifecycle.addObserver(object : DefaultLifecycleObserver{
                override fun onDestroy(owner: LifecycleOwner) {
                    context.unregisterReceiver(br)
//                    br.handler = null
                }
            })
        }
    }

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

        val vm = StateViewModel()
        vm.register(this)

        enableEdgeToEdge()
        setContent {
            TheCovenTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Greeting(
                        statusVM = vm,
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }
}

@Composable
fun Greeting(statusVM: MainActivity.StateViewModel, modifier: Modifier = Modifier) {
    val state = statusVM.state.collectAsStateWithLifecycle()
    Text(
        text = "Status: ${state.value}",
        modifier = modifier
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    TheCovenTheme {
        Greeting(MainActivity.StateViewModel())
    }
}
