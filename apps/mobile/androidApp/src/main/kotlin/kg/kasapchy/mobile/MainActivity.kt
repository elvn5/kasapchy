package kg.kasapchy.mobile

import android.app.Application
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.compose.runtime.DisposableEffect
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kg.kasapchy.mobile.app.KasapchyApp
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.DemoStore
import kg.kasapchy.mobile.features.NetworkStore
import kg.kasapchy.mobile.shared.api.MarketApi

class MarketViewModel(application: Application) : AndroidViewModel(application) {
    val store: AppStore = if (BuildConfig.DEMO_MODE) DemoStore() else NetworkStore(
        MarketApi(BuildConfig.API_BASE_URL), SecureSessionStorage(application), viewModelScope,
    )
    override fun onCleared() { (store as? NetworkStore)?.close() }
}
class MainActivity : ComponentActivity() {
    private val model by viewModels<MarketViewModel>()
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            DisposableEffect(this@MainActivity) {
                val network = model.store as? NetworkStore
                network?.socialLogin = AndroidSignIn(this@MainActivity)::login
                onDispose { network?.socialLogin = null }
            }
            BackHandler(enabled = model.store.state.canGoBack) { model.store.back() }
            KasapchyApp(model.store, BuildConfig.PRIVACY_URL, BuildConfig.DELETION_URL)
        }
    }
}
