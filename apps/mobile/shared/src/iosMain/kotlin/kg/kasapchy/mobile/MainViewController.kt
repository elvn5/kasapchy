package kg.kasapchy.mobile

import androidx.compose.ui.window.ComposeUIViewController
import kg.kasapchy.mobile.app.KasapchyApp
import kg.kasapchy.mobile.features.DemoStore

fun MainViewController() = DemoStore().let { store ->
    ComposeUIViewController { KasapchyApp(store) }
}
